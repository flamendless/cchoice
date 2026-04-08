package services

import (
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/encode"
	"cchoice/internal/logs"
	"cchoice/internal/staff"
	"cchoice/internal/utils"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type ReportService struct {
	encoder  encode.IEncode
	dbRO     database.IService
	staffLog *StaffLogsService
	holiday  *HolidayService
}

var headers = []string{
	"Date",
	"Name",
	"Time In",
	"Time Out",
	"Holiday",
	"Holiday Type",
	"Duration",
	"In Loc/Useragent",
	"Out Loc/Useragent",
	"Lunch Break In",
	"Lunch Break Out",
	"Lunch Break Duration",
	"Lunch Break In Loc/Useragent",
	"Lunch Break Out Loc/Useragent",
}

func NewReportService(
	encoder encode.IEncode,
	dbRO database.IService,
	staffLog *StaffLogsService,
	holiday *HolidayService,
) *ReportService {
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	return &ReportService{
		encoder:  encoder,
		dbRO:     dbRO,
		staffLog: staffLog,
		holiday:  holiday,
	}
}

func (s *ReportService) StreamReportCSV(
	ctx context.Context,
	writer *csv.Writer,
	data []staff.StaffRow,
	adminStaffID string,
	staffID string,
	filename string,
	startDate string,
	endDate string,
) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, adminStaffID, "export", "attendance_report_csv", result, nil); err != nil {
			logs.LogCtx(ctx).Error("[ReportService] failed to log csv report generation", zap.Error(err))
		}
	}()

	startTime, err := time.Parse(constants.DateLayoutISO, startDate)
	if err != nil {
		result = err.Error()
		return err
	}
	endTime, err := time.Parse(constants.DateLayoutISO, endDate)
	if err != nil {
		result = err.Error()
		return err
	}

	holidays, err := s.holiday.GetHolidaysByDateRange(ctx, startTime, endTime)
	if err != nil {
		result = err.Error()
		return err
	}

	holidayMap := make(map[string]Holiday, len(holidays))
	for _, h := range holidays {
		holidayMap[h.Date] = h
	}

	attMap := make(map[string]staff.StaffRow, len(data))
	for _, att := range data {
		attMap[att.ForDate] = att
	}

	allDates := make([]string, 0, len(attMap)+len(holidayMap))
	for d := range attMap {
		allDates = append(allDates, d)
	}
	for d := range holidayMap {
		if _, ok := attMap[d]; !ok {
			allDates = append(allDates, d)
		}
	}
	sort.Strings(allDates)

	if err := writer.Write([]string{"Report name: " + filename}); err != nil {
		result = err.Error()
		return err
	}
	if err := writer.Write([]string{"Start date: " + startDate}); err != nil {
		result = err.Error()
		return err
	}
	if err := writer.Write([]string{"End date: " + endDate}); err != nil {
		result = err.Error()
		return err
	}

	decodedStaffID := s.encoder.Decode(staffID)
	if decodedStaffID != encode.INVALID {
		totalDays, err := utils.GetTotalDaysBetweenDates(startDate, endDate)
		if err != nil {
			result = err.Error()
			return err
		}
		if err := writer.Write([]string{fmt.Sprintf("Total days (present/total): %d/%d", len(data), totalDays)}); err != nil {
			result = err.Error()
			return err
		}

		staffDB, err := s.dbRO.GetQueries().GetStaffByID(ctx, decodedStaffID)
		if err != nil {
			result = err.Error()
			return err
		}
		if err := writer.Write([]string{fmt.Sprintf(
			"Scheduled working time: %s - %s",
			staffDB.TimeInSchedule.String,
			staffDB.TimeOutSchedule.String,
		)}); err != nil {
			result = err.Error()
			return err
		}

		attendanceService := NewAttendanceService(s.encoder, s.dbRO, nil, s.holiday)
		extraStats := attendanceService.GetExtraStats(ctx, staffID, data)
		if err := writer.Write([]string{fmt.Sprintf("Total undertime count: %d (%.2f minutes)", extraStats.TotalUndertimeCount, extraStats.TotalUndertimeMinutes)}); err != nil {
			return err
		}
		if err := writer.Write([]string{fmt.Sprintf("Total late count: %d (%.2f minutes)", extraStats.TotalLateCount, extraStats.TotalLateMinutes)}); err != nil {
			return err
		}
		if err := writer.Write([]string{fmt.Sprintf("Total early in count: %d", extraStats.TotalEarlyInCount)}); err != nil {
			return err
		}
		if err := writer.Write([]string{fmt.Sprintf("Total overtime count: %d", extraStats.TotalOvertimeCount)}); err != nil {
			return err
		}

		if err := writer.Write(headers); err != nil {
			return err
		}
	}

	for _, dateStr := range allDates {
		if h, ok := holidayMap[dateStr]; ok {
			if att, ok := attMap[dateStr]; ok {
				if err := writer.Write(buildAttendanceRowWithHoliday(att, h)); err != nil {
					return err
				}
			} else {
				if err := writer.Write([]string{
					dateStr,
					"",
					"",
					"",
					h.Name,
					h.Type.String(),
					"",
					"",
					"",
					"",
					"",
					"",
					"",
					"",
				}); err != nil {
					return err
				}
			}
		} else if att, ok := attMap[dateStr]; ok {
			if err := writer.Write(buildAttendanceRow(att)); err != nil {
				return err
			}
		}
	}
	return nil
}

func buildAttendanceRow(att staff.StaffRow) []string {
	timeIn := utils.ExtractTimeToPH(att.TimeIn.String)
	timeOut := utils.ExtractTimeToPH(att.TimeOut.String)

	var duration string
	if att.TimeIn.Valid && att.TimeOut.Valid {
		inTime, _ := time.Parse(constants.TimeLayoutHHMMSS, timeIn)
		outTime, _ := time.Parse(constants.TimeLayoutHHMMSS, timeOut)
		duration = outTime.Sub(inTime).String()
	}

	inLocUA := formatLocationAndUseragent(att.InLocation.String, att.InBrowser, att.InBrowserVersion, att.InOs, att.InDevice)
	outLocUA := formatLocationAndUseragent(att.OutLocation.String, att.OutBrowser, att.OutBrowserVersion, att.OutOs, att.OutDevice)

	lbIn := utils.ExtractTimeToPH(att.LunchBreakIn.String)
	lbOut := utils.ExtractTimeToPH(att.LunchBreakOut.String)

	var lbDuration string
	if att.LunchBreakIn.Valid && att.LunchBreakOut.Valid {
		inTime, _ := time.Parse(constants.TimeLayoutHHMMSS, lbIn)
		outTime, _ := time.Parse(constants.TimeLayoutHHMMSS, lbOut)
		lbDuration = outTime.Sub(inTime).String()
	}

	lbInLocUA := formatLocationAndUseragent(att.LunchBreakInLocation.String, att.LunchBreakInBrowser, att.LunchBreakInBrowserVersion, att.LunchBreakInOs, att.LunchBreakInDevice)
	lbOutLocUA := formatLocationAndUseragent(att.LunchBreakOutLocation.String, att.LunchBreakOutBrowser, att.LunchBreakOutBrowserVersion, att.LunchBreakOutOs, att.LunchBreakOutDevice)

	return []string{
		att.ForDate,
		utils.BuildFullName(att.FirstName, att.MiddleName.String, att.LastName),
		timeIn,
		timeOut,
		"",
		"",
		duration,
		inLocUA,
		outLocUA,
		lbIn,
		lbOut,
		lbDuration,
		lbInLocUA,
		lbOutLocUA,
	}
}

func buildAttendanceRowWithHoliday(att staff.StaffRow, h Holiday) []string {
	timeIn := utils.ExtractTimeToPH(att.TimeIn.String)
	timeOut := utils.ExtractTimeToPH(att.TimeOut.String)

	var duration string
	if att.TimeIn.Valid && att.TimeOut.Valid {
		inTime, _ := time.Parse(constants.TimeLayoutHHMMSS, timeIn)
		outTime, _ := time.Parse(constants.TimeLayoutHHMMSS, timeOut)
		duration = outTime.Sub(inTime).String()
	}

	inLocUA := formatLocationAndUseragent(att.InLocation.String, att.InBrowser, att.InBrowserVersion, att.InOs, att.InDevice)
	outLocUA := formatLocationAndUseragent(att.OutLocation.String, att.OutBrowser, att.OutBrowserVersion, att.OutOs, att.OutDevice)

	lbIn := utils.ExtractTimeToPH(att.LunchBreakIn.String)
	lbOut := utils.ExtractTimeToPH(att.LunchBreakOut.String)

	var lbDuration string
	if att.LunchBreakIn.Valid && att.LunchBreakOut.Valid {
		inTime, _ := time.Parse(constants.TimeLayoutHHMMSS, lbIn)
		outTime, _ := time.Parse(constants.TimeLayoutHHMMSS, lbOut)
		lbDuration = outTime.Sub(inTime).String()
	}

	lbInLocUA := formatLocationAndUseragent(att.LunchBreakInLocation.String, att.LunchBreakInBrowser, att.LunchBreakInBrowserVersion, att.LunchBreakInOs, att.LunchBreakInDevice)
	lbOutLocUA := formatLocationAndUseragent(att.LunchBreakOutLocation.String, att.LunchBreakOutBrowser, att.LunchBreakOutBrowserVersion, att.LunchBreakOutOs, att.LunchBreakOutDevice)

	return []string{
		att.ForDate,
		utils.BuildFullName(att.FirstName, att.MiddleName.String, att.LastName),
		timeIn,
		timeOut,
		h.Name,
		h.Type.String(),
		duration,
		inLocUA,
		outLocUA,
		lbIn,
		lbOut,
		lbDuration,
		lbInLocUA,
		lbOutLocUA,
	}
}

func (s *ReportService) StreamReportXLSX(
	ctx context.Context,
	file *excelize.File,
	data []staff.StaffRow,
	adminStaffID string,
	staffID string,
	filename string,
	startDate string,
	endDate string,
) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, adminStaffID, "export", "attendance_report_xlsx", result, nil); err != nil {
			logs.LogCtx(ctx).Error("[ReportService] failed to log xlsx report generation", zap.Error(err))
		}
	}()

	startTime, err := time.Parse(constants.DateLayoutISO, startDate)
	if err != nil {
		result = err.Error()
		return err
	}
	endTime, err := time.Parse(constants.DateLayoutISO, endDate)
	if err != nil {
		result = err.Error()
		return err
	}

	holidays, err := s.holiday.GetHolidaysByDateRange(ctx, startTime, endTime)
	if err != nil {
		result = err.Error()
		return err
	}

	holidayMap := make(map[string]Holiday, len(holidays))
	for _, h := range holidays {
		holidayMap[h.Date] = h
	}

	attMap := make(map[string]staff.StaffRow, len(data))
	for _, att := range data {
		attMap[att.ForDate] = att
	}

	allDates := make([]string, 0, len(attMap)+len(holidayMap))
	for d := range attMap {
		allDates = append(allDates, d)
	}
	for d := range holidayMap {
		if _, ok := attMap[d]; !ok {
			allDates = append(allDates, d)
		}
	}
	sort.Strings(allDates)

	const sheet = "Sheet1"
	row := 1

	if err := file.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Report name: "+filename); err != nil {
		result = err.Error()
		return err
	}
	row++

	if err := file.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Start date: "+startDate); err != nil {
		result = err.Error()
		return err
	}
	row++

	if err := file.SetCellValue(sheet, fmt.Sprintf("A%d", row), "End date: "+endDate); err != nil {
		result = err.Error()
		return err
	}
	row++

	decodedStaffID := s.encoder.Decode(staffID)
	if decodedStaffID != encode.INVALID {
		totalDays, err := utils.GetTotalDaysBetweenDates(startDate, endDate)
		if err != nil {
			return err
		}
		if err := file.SetCellValue(
			sheet,
			fmt.Sprintf("A%d", row),
			fmt.Sprintf("Total days (present/total): %d/%d", len(data), totalDays),
		); err != nil {
			return err
		}
		row++

		staffDB, err := s.dbRO.GetQueries().GetStaffByID(ctx, decodedStaffID)
		if err != nil {
			return err
		}
		if err := file.SetCellValue(
			sheet,
			fmt.Sprintf("A%d", row),
			fmt.Sprintf(
				"Scheduled working time: %s - %s",
				staffDB.TimeInSchedule.String,
				staffDB.TimeOutSchedule.String,
			),
		); err != nil {
			return err
		}
		row++

		attendanceService := NewAttendanceService(s.encoder, s.dbRO, nil, s.holiday)
		extraStats := attendanceService.GetExtraStats(ctx, staffID, data)
		if err := file.SetCellValue(
			sheet,
			fmt.Sprintf("A%d", row),
			fmt.Sprintf("Total undertime count: %d (%.2f minutes)", extraStats.TotalUndertimeCount, extraStats.TotalUndertimeMinutes),
		); err != nil {
			return err
		}
		row++
		if err := file.SetCellValue(
			sheet,
			fmt.Sprintf("A%d", row),
			fmt.Sprintf("Total late count: %d (%.2f minutes)", extraStats.TotalLateCount, extraStats.TotalLateMinutes),
		); err != nil {
			return err
		}
		row++
		if err := file.SetCellValue(
			sheet,
			fmt.Sprintf("A%d", row),
			fmt.Sprintf("Total early in count: %d", extraStats.TotalEarlyInCount),
		); err != nil {
			return err
		}
		row++
		if err := file.SetCellValue(
			sheet,
			fmt.Sprintf("A%d", row),
			fmt.Sprintf("Total overtime count: %d", extraStats.TotalOvertimeCount),
		); err != nil {
			return err
		}
		row++

		for colIdx, header := range headers {
			col, err := excelize.ColumnNumberToName(colIdx + 1)
			if err != nil {
				return err
			}
			if err := file.SetCellValue(sheet, fmt.Sprintf("%s%d", col, row), header); err != nil {
				return err
			}
		}
		row++
	}

	for _, dateStr := range allDates {
		if h, ok := holidayMap[dateStr]; ok {
			if att, ok := attMap[dateStr]; ok {
				values := buildAttendanceRowWithHoliday(att, h)
				for colIdx, value := range values {
					col, err := excelize.ColumnNumberToName(colIdx + 1)
					if err != nil {
						return err
					}
					if err := file.SetCellValue(sheet, fmt.Sprintf("%s%d", col, row), value); err != nil {
						return err
					}
				}
			} else {
				if err := file.SetCellValue(sheet, fmt.Sprintf("A%d", row), dateStr); err != nil {
					return err
				}
				if err := file.SetCellValue(sheet, fmt.Sprintf("E%d", row), h.Name); err != nil {
					return err
				}
				if err := file.SetCellValue(sheet, fmt.Sprintf("F%d", row), h.Type.String()); err != nil {
					return err
				}
			}
			row++
		} else if att, ok := attMap[dateStr]; ok {
			values := buildAttendanceRow(att)
			for colIdx, value := range values {
				col, err := excelize.ColumnNumberToName(colIdx + 1)
				if err != nil {
					return err
				}
				if err := file.SetCellValue(sheet, fmt.Sprintf("%s%d", col, row), value); err != nil {
					return err
				}
			}
			row++
		}
	}
	return nil
}

func formatLocationAndUseragent(
	location string,
	browser, browserVersion, os, device sql.NullString,
) string {
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

func (s *ReportService) Log() {
	logs.Log().Info("[ReportService] Loaded")
}

var _ IService = (*ReportService)(nil)
