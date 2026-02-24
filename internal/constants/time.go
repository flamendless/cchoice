package constants

import "time"

var (
	DtBeginning time.Time
)

const (
	DateLayoutISO     = "2006-01-02"
	DateTimeLayoutISO = "2006-01-02 15:04:05"
	DateLayoutDisplay = "Monday, January 2, 2006"
	TimeLayoutHHMM    = "15:04"
	TimeLayoutHHMMSS  = "15:04:05"
	TimeLayoutDisplay = "03:04:05 PM"
)

func init() {
	DtBeginning = time.Unix(0, 0).UTC()
}
