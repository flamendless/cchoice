package internal

import (
	cchoice_db "cchoice/cchoice_db"
	"database/sql"
)

type AppFlags struct {
	Template               string
	Filepath               string
	Sheet                  string
	Strict                 bool
	Limit                  int
	PrintProcessedProducts bool
	VerifyPrices           bool
	DBPath                 string
	UseDB                  bool
}

type AppContext struct {
	DB      *sql.DB
	Queries *cchoice_db.Queries
}
