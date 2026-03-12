package services

import (
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/logs"
	"cchoice/internal/staff"
	"cchoice/internal/utils"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type ReportService struct {
	dbRO database.Service
}

func NewReportService(dbRO database.Service) *ReportService {
	return &ReportService{dbRO: dbRO}
}

func (s *ReportService) StreamReportCSV(
	ctx context.Context,
	writer *csv.Writer,
	data []staff.StaffRow,
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

func (s *ReportService) StreamReportXLSX(
	ctx context.Context,
	file *excelize.File,
	data []staff.StaffRow,
	filename string,
	startDate string,
	endDate string,
) error {
	row := 1

	if err := file.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), "Report name: "+filename); err != nil {
		return err
	}
	row++

	if err := file.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), "Start date: "+startDate); err != nil {
		return err
	}
	row++

	if err := file.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), "End date: "+endDate); err != nil {
		return err
	}
	row++

	headers := []string{
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
	}

	for colIdx, header := range headers {
		col, err := excelize.ColumnNumberToName(colIdx + 1)
		if err != nil {
			return err
		}
		if err := file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", col, row), header); err != nil {
			return err
		}
	}
	row++

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

		values := []string{
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
		}

		for colIdx, value := range values {
			col, err := excelize.ColumnNumberToName(colIdx + 1)
			if err != nil {
				return err
			}
			if err := file.SetCellValue("Sheet1", fmt.Sprintf("%s%d", col, row), value); err != nil {
				return err
			}
		}
		row++
	}

	return nil
}
