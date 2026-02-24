package enums

//go:generate go tool stringer -type=TimeInStatus -trimprefix=TIME_IN_STATUS_

type TimeInStatus int

const (
	TIME_IN_STATUS_UNKNOWN TimeInStatus = iota
	TIME_IN_STATUS_EARLIER
	TIME_IN_STATUS_ON_TIME
	TIME_IN_STATUS_LATE
)
