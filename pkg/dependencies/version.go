package dependencies

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"code-intelligence.com/cifuzz/pkg/log"
)

/*
Note: we made the "patch" part of the semver (when parsing the output with regex) optional
be more lenient when a command returns something like 1.2 instead of 1.2.0
*/
var (
	clangRegex = regexp.MustCompile(`(?m)clang version (?P<version>\d+\.\d+(\.\d+)?)`)
	cmakeRegex = regexp.MustCompile(`(?m)cmake version (?P<version>\d+\.\d+(\.\d+)?)`)
	llvmRegex  = regexp.MustCompile(`(?m)LLVM version (?P<version>\d+\.\d+(\.\d+)?)`)
)

type execCheck func(string, Key) (*semver.Version, error)

// small helper to reuse clang version check
func clangCheck(path string, key Key) (*semver.Version, error) {
	version, err := getVersionFromCommand(path, []string{"--version"}, clangRegex, key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return version, nil
}

// returns the currently installed clang version
func clangVersion(dep *Dependency, clangCheck execCheck) (*semver.Version, error) {
	// as we have up to three sources for a clang instance we take
	// the lowest version we find
	var minVersion *semver.Version

	// first we check if the environment variables CC and CXX are set
	// and contain a valid version number if not, we also check the
	// clang available in the path
	checkPath := false

	if cc, found := os.LookupEnv("CC"); found {
		if path.Base(cc) == "clang" {
			ccVersion, err := clangCheck(cc, dep.Key)
			if err != nil {
				return nil, err
			}
			log.Debugf("Found clang version %s in CC", ccVersion.String())
			minVersion = ccVersion
		} else {
			log.Warn("No clang found in CC")
		}
	} else {
		checkPath = true
	}

	if cxx, found := os.LookupEnv("CXX"); found {
		if path.Base(cxx) == "clang++" {
			cxxVersion, err := clangCheck(cxx, dep.Key)
			if err != nil {
				return nil, err
			}
			log.Debugf("Found clang version %s in CXX", cxxVersion.String())
			if minVersion == nil || minVersion.GreaterThan(cxxVersion) {
				minVersion = cxxVersion
			}

		} else {
			log.Warn("No clang found in CXX")
		}
	} else {
		checkPath = true
	}

	if checkPath {
		path, err := dep.finder.ClangPath()
		if err != nil {
			return nil, err
		}
		pathVersion, err := clangCheck(path, dep.Key)
		log.Debugf("Found clang version %s in PATH", pathVersion.String())
		if minVersion == nil || minVersion.GreaterThan(pathVersion) {
			minVersion = pathVersion
		}
	}

	return minVersion, nil

}

// helper for parsing the --version output for different llvm tools,
// for example llvm-cov, llvm-symbolizer
func llvmVersion(path string, dep *Dependency) (*semver.Version, error) {
	version, err := getVersionFromCommand(path, []string{"--version"}, llvmRegex, dep.Key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return version, nil
}

func cmakeVersion(dep *Dependency) (*semver.Version, error) {
	path, err := exec.LookPath("cmake")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	version, err := getVersionFromCommand(path, []string{"--version"}, cmakeRegex, dep.Key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return version, nil
}

// takes a command + args and parses the output for a semver
func getVersionFromCommand(cmdPath string, args []string, re *regexp.Regexp, key Key) (*semver.Version, error) {
	cmd := exec.Command(cmdPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return extractVersion(string(output), re, key)
}

func extractVersion(output string, re *regexp.Regexp, key Key) (*semver.Version, error) {
	result := re.FindStringSubmatch(string(output))
	if len(result) <= 1 {
		return nil, fmt.Errorf("No matching version string for %s", key)
	}

	version, err := semver.NewVersion(result[1])
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return version, nil
}
