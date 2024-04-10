package db

import (
	"cchoice/internal/logs"
	"context"
	"database/sql"

	// _ "embed"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// //go:embed sql/schema.sql
// var ddl string

func InitDB(dataSourceName string) (*sql.DB, error) {
	logs.Log().Info("Initializing DB...")
	logs.Log().Debug("opening database...")

	sqlDB, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	logs.Log().Debug("opening schema.sql")
	ddl, err := os.ReadFile("sql/schema.sql")
	if err != nil {
		return nil, err
	}

	logs.Log().Debug("executing context...")
	ctx := context.Background()
	_, err = sqlDB.ExecContext(ctx, string(ddl))
	if err != nil {
		return nil, err
	}

	// queries := db.New(sqlDB)
	// prs, err := queries.GetProducts(ctx)

	logs.Log().Info("Successfully initialized DB")

	return sqlDB, nil
}
