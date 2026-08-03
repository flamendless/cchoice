package datamigrate_test

import (
	"context"
	"testing"

	"cchoice/internal/datamigrate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScript_BuildRunArgs(t *testing.T) {
	t.Parallel()

	for _, script := range datamigrate.Scripts {
		t.Run(script.Name, func(t *testing.T) {
			t.Parallel()
			args := script.BuildRunArgs()
			require.NotEmpty(t, args)
			assert.Equal(t, script.Command, args[0])
			assert.Equal(t, "--dry-run=false", args[len(args)-1])
		})
	}

	applyDiscount := datamigrate.Scripts[0]
	assert.Equal(
		t,
		[]string{"apply_discount", "-i", "scripts/csv/sale_2025.csv", "--dry-run=false"},
		applyDiscount.BuildRunArgs(),
	)
}

func TestUp_RunsPendingScripts(t *testing.T) {
	t.Parallel()

	db := newStubDB(t, true)
	ctx := context.Background()

	_, err := db.GetDB().Exec(
		`INSERT INTO goose_db_version (version_id, is_applied) VALUES (20260710120000, 1)`,
	)
	require.NoError(t, err)

	stateBefore, err := datamigrate.GetState(ctx, db)
	require.NoError(t, err)
	require.NotEmpty(t, stateBefore.PendingScripts)

	var ran []string
	err = datamigrate.Up(ctx, db, func(_ context.Context, argv []string) error {
		ran = append(ran, argv[0])
		for _, script := range datamigrate.Scripts {
			if script.Command == argv[0] {
				return db.GetQueries().UpsertDatamigrateApplied(ctx, script.Name)
			}
		}
		return nil
	})
	require.NoError(t, err)
	assert.NotEmpty(t, ran)

	stateAfter, err := datamigrate.GetState(ctx, db)
	require.NoError(t, err)
	assert.Empty(t, stateAfter.PendingScripts)
}
