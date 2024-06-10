package ctx

import (
	cchoice_db "cchoice/cchoice_db"
	cchoicedb "cchoice/internal/cchoice_db"
	"cchoice/internal/logs"
	"database/sql"

	"go.uber.org/zap"
)

type Database struct {
	DB          *sql.DB
	DBRead      *sql.DB
	Queries     *cchoice_db.Queries
	QueriesRead *cchoice_db.Queries
}

func NewDatabaseCtx(dbPath string) *Database {
	sqlDBRead, err := cchoicedb.InitDB(dbPath, "ro")
	if err != nil {
		logs.Log().Error(
			"DB (read-only) initialization",
			zap.Error(err),
		)
		panic(err)
	}

	sqlDB, err := cchoicedb.InitDB(dbPath, "rw")
	if err != nil {
		logs.Log().Error(
			"DB (read-write) initialization",
			zap.Error(err),
		)
		panic(err)
	}

	sqlDB.SetMaxOpenConns(1)

	return &Database{
		DB:          sqlDB,
		DBRead:      sqlDBRead,
		Queries:     cchoicedb.GetQueries(sqlDB),
		QueriesRead: cchoicedb.GetQueries(sqlDBRead),
	}
}

func (ctxDB *Database) Close() {
	ctxDB.DB.Close()
	logs.Log().Info("Closed DB")
}
