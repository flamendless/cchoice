package db

import (
	"cchoice/database"
	"cchoice/internal/logs"
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func InitDB(dataSourceName string) (*sql.DB, error) {
	logs.Log().Info("Initializing DB...")
	logs.Log().Debug(
		"opening database...",
		zap.String("data source name", dataSourceName),
	)

	sqlDB, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	logs.Log().Debug("executing context...")
	ctx := context.Background()
	_, err = sqlDB.ExecContext(ctx, string(database.Schema))
	if err != nil {
		return nil, err
	}

	// queries := db.New(sqlDB)
	// prs, err := queries.GetProducts(ctx)

	logs.Log().Info("Successfully initialized DB")

	return sqlDB, nil
}
