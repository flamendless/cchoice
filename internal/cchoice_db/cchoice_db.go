package cchoicedb

import (
	"cchoice/database"
	cchoice_db "cchoice/db"
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

	logs.Log().Info("Successfully initialized DB")

	return sqlDB, nil
}

func GetQueries(sqlDB *sql.DB) *cchoice_db.Queries {
	queries := cchoice_db.New(sqlDB)
	return queries
}
