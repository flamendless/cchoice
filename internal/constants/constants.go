package constants

import "time"

var (
	DT_BEGINNING time.Time
)

func init() {
	DT_BEGINNING = time.Unix(0, 0).UTC()
}
