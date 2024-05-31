package ctx

import (
	cchoice_db "cchoice/cchoice_db"
	"database/sql"
)

type App struct {
	DB          *sql.DB
	DBRead      *sql.DB
	Queries     *cchoice_db.Queries
	QueriesRead *cchoice_db.Queries
	Metrics     *Metrics
}
