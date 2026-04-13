package services

import (
	"time"

	"database/sql"
	"cchoice/cmd/web/models"

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

type StaffDayAttendance struct {
	HasTimeIn        bool
	HasTimeOut       bool
	HasLunchBreakIn  bool
	HasLunchBreakOut bool
	Computed         *models.Attendance
	InLocation       sql.NullString
	OutLocation      sql.NullString
}
