package build

import (
	"os"
	"path"
	"runtime"

	"code-intelligence.com/cifuzz/util/envutil"
)

type Result struct {
	// Canonical path of the fuzz test executable
	Executable string
	// Canonical path of the fuzz test's default seed corpus directory
	SeedCorpus string
	// Canonical path of the build directory
	BuildDir string
	// The engine for which the fuzz test was built
	Engine string
	// The sanitizers with which the fuzz test was built
	Sanitizers []string
	// The canonical paths of the fuzz test's runtime dependencies
	RuntimeDeps []string
}

func CommonBuildEnv() ([]string, error) {
	var err error
	env := os.Environ()

	// Set CIFUZZ=1 to allow the build system to figure out that it was
	// started by cifuzz.
	env, err = envutil.Setenv(env, "CIFUZZ", "1")
	if err != nil {
		return nil, err
	}

	// On Windows, our preferred compiler is MSVC, which can't easily be run
	// from an arbitrary terminal as it requires about a dozen environment
	// variables to be set correctly. Thus, we assume users to run cifuzz from
	// a developer command prompt anyway and thus don't need to set the
	// compiler explicitly.
	if runtime.GOOS != "windows" {
		// Set the C/C++ compiler to clang/clang++ (if not already set),
		// which is needed to build a  binary with fuzzing instrumentation
		// gcc doesn't have -fsanitize=fuzzer.
		if val, ok := envutil.LookupEnv(env, "CC"); !ok || path.Base(val) != "clang" {
			env, err = envutil.Setenv(env, "CC", "clang")
			if err != nil {
				return nil, err
			}
		}
		if val, ok := envutil.LookupEnv(env, "CXX"); !ok || path.Base(val) != "clang++" {
			env, err = envutil.Setenv(env, "CXX", "clang++")
			if err != nil {
				return nil, err
			}
		}
	}

	// We don't want to fail if ASan is set up incorrectly for tools
	// built and executed during the build or they contain leaks.
	env, err = envutil.Setenv(env, "ASAN_OPTIONS", "detect_leaks=0:verify_asan_link_order=0")
	if err != nil {
		return nil, err
	}

	return env, nil
}
