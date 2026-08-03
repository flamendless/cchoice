package datamigrate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"

	"cchoice/internal/constants"
	"cchoice/internal/database"
)

type PendingGooseMigration struct {
	Version  int64
	Filename string
}

func ListPendingGooseMigrations(ctx context.Context, db database.IService) ([]PendingGooseMigration, error) {
	dir := os.Getenv("GOOSE_MIGRATION_DIR")
	if dir == "" {
		dir = "./migrations/sqlite3"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.New("failed to read goose migration dir " + dir + ": " + err.Error())
	}

	fileMigrations := make([]PendingGooseMigration, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := constants.ReGooseMigrationFile.FindStringSubmatch(entry.Name())
		if len(matches) < 2 {
			continue
		}

		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			continue
		}

		fileMigrations = append(fileMigrations, PendingGooseMigration{
			Version:  version,
			Filename: entry.Name(),
		})
	}

	slices.SortFunc(fileMigrations, func(a, b PendingGooseMigration) int {
		if a.Version < b.Version {
			return -1
		}
		if a.Version > b.Version {
			return 1
		}
		return 0
	})

	applied, err := listPendingGooseAppliedIDs(ctx, db)
	if err != nil {
		return nil, errors.New(constants.MsgFailedReadGooseVersion + ": " + err.Error())
	}

	pending := make([]PendingGooseMigration, 0)
	for _, migration := range fileMigrations {
		if _, ok := applied[migration.Version]; ok {
			continue
		}
		pending = append(pending, migration)
	}

	return pending, nil
}

func formatPendingGooseLine(order int, migration PendingGooseMigration) string {
	return fmt.Sprintf(
		"%d. %d (%s) %s",
		order,
		migration.Version,
		FormatGooseVersion(migration.Version),
		migration.Filename,
	)
}
