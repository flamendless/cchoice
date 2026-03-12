package services

import (
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type StaffRow struct {
	ID                          int64
	StaffID                     int64
	ForDate                     string
	TimeIn                      sql.NullString
	TimeOut                     sql.NullString
	InLocation                  sql.NullString
	OutLocation                 sql.NullString
	InUseragentID               sql.NullInt64
	OutUseragentID              sql.NullInt64
	LunchBreakIn                sql.NullString
	LunchBreakOut               sql.NullString
	LunchBreakInLocation        sql.NullString
	LunchBreakOutLocation       sql.NullString
	LunchBreakInUseragentID     sql.NullInt64
	LunchBreakOutUseragentID    sql.NullInt64
	CreatedAt                   string
	UpdatedAt                   string
	FirstName                   string
	MiddleName                  sql.NullString
	LastName                    string
	InBrowser                   sql.NullString
	InBrowserVersion            sql.NullString
	InOs                        sql.NullString
	InDevice                    sql.NullString
	OutBrowser                  sql.NullString
	OutBrowserVersion           sql.NullString
	OutOs                       sql.NullString
	OutDevice                   sql.NullString
	LunchBreakInBrowser         sql.NullString
	LunchBreakInBrowserVersion  sql.NullString
	LunchBreakInOs              sql.NullString
	LunchBreakInDevice          sql.NullString
	LunchBreakOutBrowser        sql.NullString
	LunchBreakOutBrowserVersion sql.NullString
	LunchBreakOutOs             sql.NullString
	LunchBreakOutDevice         sql.NullString
}

type ReportService struct {
	dbRO database.Service
}

func NewReportService(dbRO database.Service) *ReportService {
	return &ReportService{dbRO: dbRO}
}

func (s *ReportService) GetAttendanceReport(
	ctx context.Context,
	staffID int64,
	startDate string,
	endDate string,
) ([]StaffRow, error) {
	var data []StaffRow
	if staffID != encode.INVALID {
		attendances, err := s.dbRO.GetQueries().GetStaffAttendanceByDateRangeAndStaffID(ctx, queries.GetStaffAttendanceByDateRangeAndStaffIDParams{
			StaffID:   staffID,
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			return data, err
		}

		data = make([]StaffRow, 0, len(attendances))
		for _, att := range attendances {
			data = append(data, StaffRow(att))
		}
	} else {
		attendances, err := s.dbRO.GetQueries().GetStaffAttendanceByDateRange(ctx, queries.GetStaffAttendanceByDateRangeParams{
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			return data, err
		}

		data = make([]StaffRow, 0, len(attendances))
		for _, att := range attendances {
			data = append(data, StaffRow(att))
		}
	}
	return data, nil
}

func (s *ReportService) StreamReport(
	ctx context.Context,
	writer *csv.Writer,
	data []StaffRow,
	filename string,
	startDate string,
	endDate string,
) error {
	if err := writer.Write([]string{"Report name: " + filename}); err != nil {
		return err
	}

	if err := writer.Write([]string{"Start date: " + startDate}); err != nil {
		return err
	}
	if err := writer.Write([]string{"End date: " + endDate}); err != nil {
		return err
	}
	if err := writer.Write([]string{
		"date",
		"name of staff",
		"time in",
		"time out",
		"duration",
		"in location and useragent",
		"out location and useragent",
		"lunch break start",
		"lunch break end",
		"lunch break duration",
		"lunch break start location and useragent",
		"lunch break end location and useragent",
	}); err != nil {
		return err
	}

	for _, att := range data {
		timeIn := utils.ExtractTimeToPH(att.TimeIn.String)
		timeOut := utils.ExtractTimeToPH(att.TimeOut.String)

		var duration string
		if att.TimeIn.Valid && att.TimeOut.Valid {
			inTime, err := time.Parse(constants.TimeLayoutHHMMSS, timeIn)
			if err != nil {
				logs.Log().Warn("report generation", zap.String("time in", timeIn), zap.Error(err))
			}

			outTime, err := time.Parse(constants.TimeLayoutHHMMSS, timeOut)
			if err != nil {
				logs.Log().Warn("report generation", zap.String("time out", timeOut), zap.Error(err))
			}

			duration = outTime.Sub(inTime).String()
		}

		inLocUA := formatLocationAndUseragent(att.InLocation.String, att.InBrowser, att.InBrowserVersion, att.InOs, att.InDevice)
		outLocUA := formatLocationAndUseragent(att.OutLocation.String, att.OutBrowser, att.OutBrowserVersion, att.OutOs, att.OutDevice)

		lbIn := utils.ExtractTimeToPH(att.LunchBreakIn.String)
		lbOut := utils.ExtractTimeToPH(att.LunchBreakOut.String)

		var lbDuration string
		if att.LunchBreakIn.Valid && att.LunchBreakOut.Valid {
			inTime, err := time.Parse(constants.TimeLayoutHHMMSS, lbIn)
			if err != nil {
				logs.Log().Warn("report generation", zap.String("lunch break start", lbIn), zap.Error(err))
			}

			outTime, err := time.Parse(constants.TimeLayoutHHMMSS, lbOut)
			if err != nil {
				logs.Log().Warn("report generation", zap.String("lunch break end", lbOut), zap.Error(err))
			}

			lbDuration = outTime.Sub(inTime).String()
		}

		lbInLocUA := formatLocationAndUseragent(att.LunchBreakInLocation.String, att.LunchBreakInBrowser, att.LunchBreakInBrowserVersion, att.LunchBreakInOs, att.LunchBreakInDevice)
		lbOutLocUA := formatLocationAndUseragent(att.LunchBreakOutLocation.String, att.LunchBreakOutBrowser, att.LunchBreakOutBrowserVersion, att.LunchBreakOutOs, att.LunchBreakOutDevice)

		if err := writer.Write([]string{
			att.ForDate,
			utils.BuildFullName(att.FirstName, att.MiddleName.String, att.LastName),
			timeIn,
			timeOut,
			duration,
			inLocUA,
			outLocUA,
			lbIn,
			lbOut,
			lbDuration,
			lbInLocUA,
			lbOutLocUA,
		}); err != nil {
			return err
		}
	}
	return nil
}

func formatLocationAndUseragent(location string, browser, browserVersion, os, device sql.NullString) string {
	var result string
	if location != "" {
		if lat, long, ok := utils.ParseLocation(sql.NullString{
			String: location,
			Valid:  true,
		}); ok {
			result = fmt.Sprintf("(%f,%f)", lat, long)
		}
	}
	if browser.Valid && browser.String != "" {
		if result != "" {
			result += " - "
		}
		result += fmt.Sprintf("%s/%s/%s", browser.String, os.String, device.String)
	}
	return result
}
