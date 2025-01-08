package models

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
