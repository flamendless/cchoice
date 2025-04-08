package constants

import "time"

var (
	DtBeginning time.Time
)

func init() {
	DtBeginning = time.Unix(0, 0).UTC()
}
