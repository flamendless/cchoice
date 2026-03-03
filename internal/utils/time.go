package utils

import (
	"fmt"
	"time"

	"cchoice/internal/constants"
)

var phLocation *time.Location

func init() {
	var err error
	phLocation, err = time.LoadLocation("Asia/Manila")
	if err != nil {
		phLocation = time.FixedZone("PHT", 8*60*60)
	}
}

func NowPH() time.Time {
	return time.Now().In(phLocation)
}

func TimeToMinutes(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	layouts := []string{
		constants.DateTimeLayoutISO,
		constants.TimeLayoutHHMMSS,
		constants.TimeLayoutHHMM,
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t.Hour()*60 + t.Minute(), true
		}
	}
	return 0, false
}

func FormatDurationFromMinutes(m int) string {
	if m < 0 {
		return "-"
	}
	h := m / 60
	min := m % 60
	return fmt.Sprintf("%dh %dm", h, min)
}

func ExtractTime(datetimeStr string) string {
	if datetimeStr == "" {
		return ""
	}
	t, err := time.Parse("2006-01-02 15:04:05", datetimeStr)
	if err != nil {
		return datetimeStr
	}
	return t.Format("15:04:05")
}

func ParseAttendanceDate(date string) string {
	if date == "" {
		return time.Now().Format(constants.DateLayoutISO)
	}
	if _, err := time.Parse(constants.DateLayoutISO, date); err != nil {
		return time.Now().Format(constants.DateLayoutISO)
	}
	return date
}
