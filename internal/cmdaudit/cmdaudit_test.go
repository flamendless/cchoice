package cmdaudit

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldSkipCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		command  string
		parent   string
		expected bool
	}{
		{name: "nil", command: "", parent: "", expected: true},
		{name: "empty", command: "", parent: "root", expected: true},
		{name: "web", command: "web", expected: true},
		{name: "api", command: "api", expected: true},
		{name: "datamigrate", command: "datamigrate", expected: true},
		{name: "datamigrate check child", command: "check", parent: "datamigrate", expected: true},
		{name: "populate", command: "populate_brand_slugs", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var cmd *cobra.Command
			if tt.name == "nil" {
				assert.True(t, ShouldSkipCommand(nil))
				return
			}
			cmd = &cobra.Command{Use: tt.command}
			if tt.parent != "" {
				parent := &cobra.Command{Use: tt.parent}
				parent.AddCommand(cmd)
			}
			assert.Equal(t, tt.expected, ShouldSkipCommand(cmd))
		})
	}
}

func TestIsSystemCLIRun(t *testing.T) {
	t.Setenv(EnvCLIStaffID, "")
	assert.True(t, isSystemCLIRun())

	t.Setenv(EnvCLIStaffID, "encoded-staff-id")
	assert.False(t, isSystemCLIRun())

	t.Setenv(EnvCLIStaffID, "   ")
	assert.True(t, isSystemCLIRun())
}

func TestShouldRedactFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flag     string
		expected bool
	}{
		{name: "password", flag: "password", expected: true},
		{name: "api key", flag: "api_key", expected: true},
		{name: "secret token", flag: "client_secret", expected: true},
		{name: "dry run", flag: "dry-run", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, shouldRedactFlag(tt.flag))
		})
	}
}

func TestCollectSanitizedFlags(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("dry-run", "true", "")
	cmd.Flags().String("password", "secret", "")
	require.NoError(t, cmd.Flags().Set("dry-run", "false"))
	require.NoError(t, cmd.Flags().Set("password", "secret"))

	flags := collectSanitizedFlags(cmd)
	assert.Equal(t, "false", flags["dry-run"])
	assert.NotContains(t, flags, "password")
}

func TestBuildResultSuccess(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "populate_brand_slugs"}
	cmd.Flags().Bool("dry-run", true, "")
	require.NoError(t, cmd.Flags().Set("dry-run", "false"))

	extra := &resultData{
		summary: "updated=3",
		details: map[string]any{"failures": 0},
	}

	payload, err := buildResult(cmd, []string{"arg1"}, nil, 150*time.Millisecond, extra)
	require.NoError(t, err)

	var result auditResult
	require.NoError(t, json.Unmarshal([]byte(payload), &result))

	assert.Equal(t, "success", result.Status)
	assert.Equal(t, "populate_brand_slugs", result.Command)
	assert.Equal(t, []string{"arg1"}, result.Args)
	assert.Equal(t, "false", result.Flags["dry-run"])
	assert.Equal(t, int64(150), result.DurationMS)
	assert.Equal(t, "updated=3", result.Summary)
	assert.Equal(t, float64(0), result.Details["failures"])
}

func TestBuildResultError(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "seed_holidays"}
	payload, err := buildResult(cmd, nil, errors.New("db unavailable"), time.Second, nil)
	require.NoError(t, err)

	var result auditResult
	require.NoError(t, json.Unmarshal([]byte(payload), &result))

	assert.Equal(t, "error", result.Status)
	assert.Equal(t, "db unavailable", result.Error)
}

func TestTruncateError(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", maxResultErrorLen+50)
	truncated := truncateError(long)
	assert.LessOrEqual(t, len(truncated), maxResultErrorLen+3)
	assert.True(t, strings.HasSuffix(truncated, "..."))
}

func TestSetSummaryStoresInContext(t *testing.T) {
	t.Parallel()

	ctx := SetSummary(t.Context(), "updated=1")
	data := resultFromContext(ctx)
	assert.Equal(t, "updated=1", data.summary)
}

func TestCollectSanitizedFlagsNilCommand(t *testing.T) {
	t.Parallel()

	flags := collectSanitizedFlags(nil)
	assert.Empty(t, flags)
}

func TestCollectSanitizedFlagsPersistentFlags(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "test"}
	cmd.PersistentFlags().AddFlagSet(pflag.NewFlagSet("persistent", pflag.ContinueOnError))
	cmd.PersistentFlags().String("token", "", "")
	require.NoError(t, cmd.PersistentFlags().Set("token", "abc"))

	flags := collectSanitizedFlags(cmd)
	assert.Empty(t, flags)
}
