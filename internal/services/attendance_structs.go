package services

import (
	"time"

	"cchoice/internal/enums"
)

type attendanceStatusResult struct {
	duration      string
	durationColor string
	inStatus      enums.TimeInStatus
	outStatus     enums.TimeOutStatus
	inLate        time.Duration
	undertime     time.Duration
	earlyIn       time.Duration
}

type AttendanceExtraStats struct {
	TotalUndertimeMinutes float64
	TotalLateMinutes      float64
	TotalUndertimeCount   int
	TotalLateCount        int
	TotalEarlyInCount     int
	TotalOvertimeCount    int
}
