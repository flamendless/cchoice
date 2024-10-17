package constants

import "time"

var (
	DT_BEGINNING time.Time
)

const (
	DEC_SCALE = 10
)

func init() {
	DT_BEGINNING = time.Unix(0, 0).UTC()
}
