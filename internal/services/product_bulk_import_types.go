package services

type BulkImportRowError struct {
	Line   int
	Serial string
	Reason string
}

type BulkImportResult struct {
	Created int
	Updated int
	Failed  int
	Errors  []BulkImportRowError
}

type BulkImportFieldChange struct {
	Field  string
	Before string
	After  string
}

type BulkImportPreviewRow struct {
	Line        int
	Serial      string
	ProductName string
	Action      string
	Error       string
	Changes     []BulkImportFieldChange
	Selectable  bool
}

type BulkImportPreview struct {
	Rows              []BulkImportPreviewRow
	TotalRows         int
	SelectableCount   int
	CreateCount       int
	UpdateCount       int
	UnchangedCount    int
	ErrorCount        int
}

type ProductImportSessionData struct {
	Headers []string
	Records [][]string
}
