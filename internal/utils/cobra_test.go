package utils

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDryRun(t *testing.T) {
	t.Parallel()

	cmdWithDryRun := &cobra.Command{Use: "test"}
	cmdWithDryRun.Flags().Bool("dry-run", true, "")
	require.NoError(t, cmdWithDryRun.Flags().Set("dry-run", "true"))
	assert.True(t, IsDryRun(cmdWithDryRun))

	require.NoError(t, cmdWithDryRun.Flags().Set("dry-run", "false"))
	assert.False(t, IsDryRun(cmdWithDryRun))

	cmdWithout := &cobra.Command{Use: "test"}
	assert.False(t, IsDryRun(cmdWithout))
}
