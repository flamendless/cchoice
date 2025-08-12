package models

import "cchoice/internal/database"

type ParseProducts struct {
	DB      database.Service
	Metrics *Metrics
}

type ParseProductsFlags struct {
	Template               string
	Filepath               string
	Sheet                  string
	DBPath                 string
	ImagesBasePath         string
	ImagesFormat           string
	Limit                  int
	Strict                 bool
	PrintProcessedProducts bool
	VerifyPrices           bool
	UseDB                  bool
	PanicOnFirstDBError    bool
}
