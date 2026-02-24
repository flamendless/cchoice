package enums

//go:generate go tool stringer -type=TimeOutStatus -trimprefix=TIME_OUT_STATUS_

type TimeOutStatus int

const (
	TIME_OUT_STATUS_UNKNOWN TimeOutStatus = iota
	TIME_OUT_STATUS_UNDERTIME
	TIME_OUT_STATUS_ON_TIME
	TIME_OUT_STATUS_OVERTIME
)
