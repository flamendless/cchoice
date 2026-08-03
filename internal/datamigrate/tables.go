package datamigrate

import (
	"context"

	"cchoice/internal/database"
)

func datamigrateTableExists(ctx context.Context, db database.IService) (bool, error) {
	return tableExists(ctx, db, "tbl_datamigrate_applied")
}

func gooseTableExists(ctx context.Context, db database.IService) (bool, error) {
	return tableExists(ctx, db, "goose_db_version")
}

func tableExists(ctx context.Context, db database.IService, tableName string) (bool, error) {
	var count int64
	if err := db.GetDB().QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?",
		tableName,
	).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func GooseMaxAppliedVersion(ctx context.Context, db database.IService) (int64, error) {
	exists, err := gooseTableExists(ctx, db)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	version, err := db.GetQueries().GetGooseMaxAppliedVersion(ctx)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func listPendingGooseAppliedIDs(ctx context.Context, db database.IService) (map[int64]struct{}, error) {
	exists, err := gooseTableExists(ctx, db)
	if err != nil {
		return nil, err
	}
	if !exists {
		return map[int64]struct{}{}, nil
	}

	ids, err := db.GetQueries().ListGooseAppliedVersionIDs(ctx)
	if err != nil {
		return nil, err
	}

	applied := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		applied[id] = struct{}{}
	}
	return applied, nil
}
