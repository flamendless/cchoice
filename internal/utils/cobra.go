package utils

import "github.com/spf13/cobra"

func IsDryRun(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}

	if cmd.Flags().Lookup("dry-run") == nil {
		return false
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return false
	}
	return dryRun
}
