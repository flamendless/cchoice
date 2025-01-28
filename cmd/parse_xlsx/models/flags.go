package models

import "cchoice/internal/database"

type ParseXLSX struct {
	DB      database.Service
	Metrics *Metrics
}

type ParseXLSXFlags struct {
	Template               string
	Filepath               string
	Sheet                  string
	DBPath                 string
	ImagesBasePath         string
	Limit                  int
	Strict                 bool
	PrintProcessedProducts bool
	VerifyPrices           bool
	UseDB                  bool
	PanicOnFirstDBError    bool
}
