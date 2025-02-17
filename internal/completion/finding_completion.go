package completion

import (
	"github.com/spf13/cobra"

	"code-intelligence.com/cifuzz/internal/config"
	"code-intelligence.com/cifuzz/pkg/cmdutils"
	"code-intelligence.com/cifuzz/pkg/finding"
	"code-intelligence.com/cifuzz/pkg/log"
)

// ValidFindings can be used as a cobra ValidArgsFunction that completes
// finding names.
func ValidFindings(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Change the directory if the `--directory` flag was set
	err := cmdutils.Chdir()
	if err != nil {
		log.Error(err, err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	// Find the project directory
	projectDir, err := config.FindProjectDir()
	if err != nil {
		log.Error(err, err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	findings, err := finding.ListFindings(projectDir)
	if err != nil {
		log.Error(err, err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	var findingNames []string
	for _, f := range findings {
		findingNames = append(findingNames, f.Name)
	}
	return findingNames, cobra.ShellCompDirectiveNoFileComp
}
