package datamigrate_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cchoice/internal/datamigrate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListPendingGooseMigrations(t *testing.T) {
	migrationDir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(migrationDir, "20260109164336_applied.sql"),
		[]byte("-- applied"),
		0o644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(migrationDir, "20260803140000_pending.sql"),
		[]byte("-- pending"),
		0o644,
	))

	t.Setenv("GOOSE_MIGRATION_DIR", migrationDir)

	db := newStubDB(t, false)
	ctx := context.Background()

	_, err := db.GetDB().Exec(
		`INSERT INTO goose_db_version (version_id, is_applied) VALUES (20260109164336, 1)`,
	)
	require.NoError(t, err)

	pending, err := datamigrate.ListPendingGooseMigrations(ctx, db)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Equal(t, int64(20260803140000), pending[0].Version)
	assert.Equal(t, "20260803140000_pending.sql", pending[0].Filename)
}

func TestListPendingGooseMigrationsAllPendingWhenGooseTableMissing(t *testing.T) {
	migrationDir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(migrationDir, "20260109164336_first.sql"),
		[]byte("-- first"),
		0o644,
	))

	t.Setenv("GOOSE_MIGRATION_DIR", migrationDir)

	db := newStubDB(t, false)
	ctx := context.Background()

	pending, err := datamigrate.ListPendingGooseMigrations(ctx, db)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Equal(t, "20260109164336_first.sql", pending[0].Filename)
}
