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
	Strict                 bool
	Limit                  int
	PrintProcessedProducts bool
	VerifyPrices           bool
	DBPath                 string
	UseDB                  bool
	PanicOnFirstDBError    bool
	ImagesBasePath         string
}
