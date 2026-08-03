package datamigrate_test

import (
	"context"
	"database/sql"
	"testing"

	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/datamigrate"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubDB struct {
	queries *queries.Queries
	db      *sql.DB
}

func (s stubDB) Health() map[string]string { return nil }

func (s stubDB) Close() error { return s.db.Close() }

func (s stubDB) GetQueries() *queries.Queries { return s.queries }

func (s stubDB) GetDB() *sql.DB { return s.db }

func newStubDB(t *testing.T, withDatamigrateTable bool) database.IService {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE goose_db_version (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version_id INTEGER NOT NULL,
			is_applied INTEGER NOT NULL,
			tstamp TIMESTAMP DEFAULT (datetime('now'))
		)
	`)
	require.NoError(t, err)

	if withDatamigrateTable {
		_, err = db.Exec(`
			CREATE TABLE tbl_datamigrate_applied (
				name TEXT PRIMARY KEY,
				applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
			)
		`)
		require.NoError(t, err)
	}

	return stubDB{
		queries: queries.New(db),
		db:      db,
	}
}

func TestMatchScript(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "apply_discount"}
	cmd.Flags().String("input_file", "", "")
	require.NoError(t, cmd.Flags().Set("input_file", "scripts/csv/sale_2025.csv"))

	script, ok := datamigrate.MatchScript(cmd, nil)
	require.True(t, ok)
	assert.Equal(t, "apply_discount:sale_2025", script.Name)

	cmdSlug := &cobra.Command{Use: "populate_product_slugs"}
	script, ok = datamigrate.MatchScript(cmdSlug, nil)
	require.True(t, ok)
	assert.Equal(t, "populate_product_slugs", script.Name)

	cmdOther, ok := datamigrate.MatchScript(&cobra.Command{Use: "web"}, nil)
	assert.False(t, ok)
	assert.Nil(t, cmdOther)
}

func TestGetStatePendingMigration(t *testing.T) {
	t.Parallel()

	db := newStubDB(t, false)
	state, err := datamigrate.GetState(context.Background(), db)
	require.NoError(t, err)
	assert.True(t, state.PendingMigration)
	assert.Empty(t, state.PendingScripts)
}

func TestGetStatePendingScripts(t *testing.T) {
	t.Parallel()

	db := newStubDB(t, true)
	ctx := context.Background()

	_, err := db.GetDB().Exec(
		`INSERT INTO goose_db_version (version_id, is_applied) VALUES (20260710120000, 1)`,
	)
	require.NoError(t, err)

	state, err := datamigrate.GetState(ctx, db)
	require.NoError(t, err)
	assert.False(t, state.PendingMigration)

	names := make(map[string]struct{}, len(state.PendingScripts))
	for _, p := range state.PendingScripts {
		names[p.Name] = struct{}{}
	}

	assert.Contains(t, names, "populate_product_slugs")
	assert.Contains(t, names, "populate_brand_slugs")
	assert.Contains(t, names, "apply_discount:sale_2025")
	assert.NotContains(t, names, "populate_product_images_cdn")

	assert.Equal(t, 1, state.PendingScripts[0].Order)
	assert.Equal(t, "apply_discount:sale_2025", state.PendingScripts[0].Name)
	assert.Equal(t, 2, state.PendingScripts[1].Order)
	assert.Equal(t, "populate_product_slugs", state.PendingScripts[1].Name)
	assert.Equal(t, 3, state.PendingScripts[2].Order)
	assert.Equal(t, "populate_brand_slugs", state.PendingScripts[2].Name)
}

func TestTryRecordSkipsDryRun(t *testing.T) {
	t.Parallel()

	db := newStubDB(t, true)
	cmd := &cobra.Command{Use: "populate_product_slugs"}
	cmd.Flags().Bool("dry-run", true, "")
	require.NoError(t, cmd.Flags().Set("dry-run", "true"))

	err := datamigrate.TryRecord(context.Background(), db, cmd, nil, nil)
	require.NoError(t, err)

	applied, err := db.GetQueries().ListDatamigrateAppliedNames(context.Background())
	require.NoError(t, err)
	assert.Empty(t, applied)
}

func TestTryRecordApplied(t *testing.T) {
	t.Parallel()

	db := newStubDB(t, true)
	cmd := &cobra.Command{Use: "populate_brand_slugs"}
	cmd.Flags().Bool("dry-run", true, "")
	require.NoError(t, cmd.Flags().Set("dry-run", "false"))

	err := datamigrate.TryRecord(context.Background(), db, cmd, nil, nil)
	require.NoError(t, err)

	applied, err := db.GetQueries().ListDatamigrateAppliedNames(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"populate_brand_slugs"}, applied)
}

func TestCheckPendingScriptsError(t *testing.T) {
	t.Parallel()

	db := newStubDB(t, true)
	ctx := context.Background()

	_, err := db.GetDB().Exec(
		`INSERT INTO goose_db_version (version_id, is_applied) VALUES (20260710120000, 1)`,
	)
	require.NoError(t, err)

	err = datamigrate.Check(ctx, db)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pending post-migrate scripts")
}
