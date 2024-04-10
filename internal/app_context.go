package internal

import "database/sql"

type AppFlags struct {
	Template               string
	Filepath               string
	Sheet                  string
	Strict                 bool
	Limit                  int
	PrintProcessedProducts bool
	DBPath                 string
	UseDB                  bool
}

type AppContext struct {
	DB *sql.DB
}
