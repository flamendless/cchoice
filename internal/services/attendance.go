package services

import (
	"context"
	"database/sql"
	"time"

	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/staff"
	"cchoice/internal/types"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

type AttendanceService struct {
	encoder      encode.IEncode
	dbRO         database.Service
	dbRW         database.Service
	shopLocation types.Location
}

type attendanceStatusResult struct {
	inStatus      enums.TimeInStatus
	outStatus     enums.TimeOutStatus
	duration      string
	durationColor string
	inLate        time.Duration
	undertime     time.Duration
	earlyIn       time.Duration
}

func NewAttendanceService(
	encoder encode.IEncode,
	ro, rw database.Service,
) *AttendanceService {
	return &AttendanceService{
		encoder:      encoder,
		dbRO:         ro,
		dbRW:         rw,
		shopLocation: conf.Conf().Settings.ShopLocation,
	}
}

func (s *AttendanceService) TimeIn(
	ctx context.Context,
	staffID string,
	date string,
	now string,
	location sql.NullString,
	useragentID sql.NullInt64,
) error {
	dbStaffID := s.encoder.Decode(staffID)
	existing, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx,
		queries.GetStaffAttendanceByDateParams{
			StaffID: dbStaffID,
			ForDate: date,
		})

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == nil && existing.TimeIn.Valid {
		return sql.ErrNoRows
	}

	if err == sql.ErrNoRows {
		_, err = s.dbRW.GetQueries().CreateStaffAttendance(ctx,
			queries.CreateStaffAttendanceParams{
				StaffID:       dbStaffID,
				ForDate:       date,
				TimeIn:        sql.NullString{String: now, Valid: true},
				TimeOut:       sql.NullString{},
				InLocation:    location,
				InUseragentID: useragentID,
			})
		return err
	}

	_, err = s.dbRW.GetQueries().UpdateStaffAttendanceTimeIn(ctx,
		queries.UpdateStaffAttendanceTimeInParams{
			StaffID:       dbStaffID,
			ForDate:       date,
			TimeIn:        sql.NullString{String: now, Valid: true},
			InLocation:    location,
			InUseragentID: useragentID,
		})

	return err
}

func (s *AttendanceService) TimeOut(
	ctx context.Context,
	staffID string,
	date string,
	now string,
	location sql.NullString,
	useragentID sql.NullInt64,
) error {
	dbStaffID := s.encoder.Decode(staffID)
	existing, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx,
		queries.GetStaffAttendanceByDateParams{
			StaffID: dbStaffID,
			ForDate: date,
		})
	if err != nil {
		return err
	}

	if !existing.TimeIn.Valid {
		return sql.ErrNoRows
	}

	if existing.TimeOut.Valid {
		return sql.ErrTxDone
	}

	_, err = s.dbRW.GetQueries().UpdateStaffAttendanceTimeOut(ctx,
		queries.UpdateStaffAttendanceTimeOutParams{
			TimeOut:        sql.NullString{String: now, Valid: true},
			OutLocation:    location,
			OutUseragentID: useragentID,
			StaffID:        dbStaffID,
			ForDate:        date,
		})

	return err
}

func (s *AttendanceService) LunchBreakIn(
	ctx context.Context,
	staffID string,
	date string,
	now string,
	location sql.NullString,
	useragentID sql.NullInt64,
) error {
	_, err := s.dbRW.GetQueries().UpdateStaffAttendanceLunchBreakIn(ctx, queries.UpdateStaffAttendanceLunchBreakInParams{
		LunchBreakIn:            sql.NullString{String: now, Valid: true},
		LunchBreakInLocation:    location,
		LunchBreakInUseragentID: useragentID,
		StaffID:                 s.encoder.Decode(staffID),
		ForDate:                 date,
	})
	return err
}

func (s *AttendanceService) LunchBreakOut(
	ctx context.Context,
	staffID string,
	date string,
	now string,
	location sql.NullString,
	useragentID sql.NullInt64,
) error {
	dbStaffID := s.encoder.Decode(staffID)
	existing, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: dbStaffID,
		ForDate: date,
	})
	if err != nil {
		return err
	}
	if !existing.LunchBreakIn.Valid {
		return sql.ErrNoRows
	}
	if existing.LunchBreakOut.Valid {
		return sql.ErrTxDone
	}

	_, err = s.dbRW.GetQueries().UpdateStaffAttendanceLunchBreakOut(ctx, queries.UpdateStaffAttendanceLunchBreakOutParams{
		LunchBreakOut:            sql.NullString{String: now, Valid: true},
		LunchBreakOutLocation:    location,
		LunchBreakOutUseragentID: useragentID,
		StaffID:                  dbStaffID,
		ForDate:                  date,
	})
	return err
}

func (s *AttendanceService) TimeOff(
	ctx context.Context,
	staffID string,
	timeOffType enums.TimeOff,
	description string,
	startDate time.Time,
	endDate time.Time,
	useragentID sql.NullInt64,
) error {
	_, err := s.dbRW.GetQueries().CreateStaffTimeOff(
		ctx,
		queries.CreateStaffTimeOffParams{
			Type:        timeOffType.String(),
			StartDate:   startDate,
			EndDate:     endDate,
			Description: description,
			StaffID:     s.encoder.Decode(staffID),
			UseragentID: useragentID,
		},
	)
	return err
}

func (s *AttendanceService) UpsertLocation(
	ctx context.Context,
	staffID string,
	date string,
	location sql.NullString,
) error {
	dbStaffID := s.encoder.Decode(staffID)
	_, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: dbStaffID,
		ForDate: date,
	})

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		_, err = s.dbRW.GetQueries().CreateStaffAttendance(ctx, queries.CreateStaffAttendanceParams{
			StaffID:     dbStaffID,
			ForDate:     date,
			TimeIn:      sql.NullString{},
			TimeOut:     sql.NullString{},
			OutLocation: location,
		})
		return err
	}

	_, err = s.dbRW.GetQueries().UpdateStaffAttendanceLocation(ctx, queries.UpdateStaffAttendanceLocationParams{
		OutLocation: location,
		StaffID:     dbStaffID,
		ForDate:     date,
	})

	return err
}

func (s *AttendanceService) ApproveTimeOff(
	ctx context.Context,
	timeOffID string,
	approvedByID string,
) error {
	_, err := s.dbRW.GetQueries().ApproveStaffTimeOff(ctx, queries.ApproveStaffTimeOffParams{
		ApprovedBy: sql.NullInt64{Int64: s.encoder.Decode(approvedByID), Valid: true},
		ID:         s.encoder.Decode(timeOffID),
	})
	return err
}

func (s *AttendanceService) CancelTimeOff(ctx context.Context, timeOffID string) error {
	_, err := s.dbRW.GetQueries().CancelStaffTimeOff(ctx, s.encoder.Decode(timeOffID))
	return err
}

func (s *AttendanceService) GetAttendance(
	ctx context.Context,
	staffID string,
	startDate string,
	endDate string,
) ([]staff.StaffRow, error) {
	dbStaffID := s.encoder.Decode(staffID)
	var data []staff.StaffRow
	if dbStaffID != encode.INVALID {
		attendances, err := s.dbRO.GetQueries().GetStaffAttendanceByDateRangeAndStaffID(ctx, queries.GetStaffAttendanceByDateRangeAndStaffIDParams{
			StaffID:   dbStaffID,
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			return data, err
		}

		data = make([]staff.StaffRow, 0, len(attendances))
		for _, att := range attendances {
			data = append(data, staff.StaffRow(att))
		}
	} else {
		attendances, err := s.dbRO.GetQueries().GetStaffAttendanceByDateRange(ctx, queries.GetStaffAttendanceByDateRangeParams{
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			return data, err
		}

		data = make([]staff.StaffRow, 0, len(attendances))
		for _, att := range attendances {
			data = append(data, staff.StaffRow(att))
		}
	}
	return data, nil
}

func (s *AttendanceService) ComputeAllAttendanceData(
	staffs []queries.GetAllStaffsRow,
	attendances []staff.StaffRow,
) []models.Attendance {
	staffMap := make(map[int64]queries.GetAllStaffsRow)
	for _, staff := range staffs {
		staffMap[staff.ID] = staff
	}

	attendanceData := make([]models.Attendance, 0, len(attendances))
	for _, att := range attendances {
		st, ok := staffMap[att.StaffID]
		if !ok {
			continue
		}

		staffBase := staff.StaffRowBase{
			FirstName:       st.FirstName,
			MiddleName:      st.MiddleName,
			LastName:        st.LastName,
			TimeInSchedule:  st.TimeInSchedule,
			TimeOutSchedule: st.TimeOutSchedule,
		}

		attendanceData = append(
			attendanceData,
			s.ComputeData(staffBase, att),
		)
	}
	return attendanceData
}

func (s *AttendanceService) ComputeData(
	staff staff.StaffRowBase,
	att staff.StaffRow,
) models.Attendance {
	schedIn, schedOut := "", ""
	if staff.TimeInSchedule.Valid {
		schedIn = staff.TimeInSchedule.String
	}
	if staff.TimeOutSchedule.Valid {
		schedOut = staff.TimeOutSchedule.String
	}
	timeIn, timeOut := utils.ExtractTimeToPH(att.TimeIn.String), utils.ExtractTimeToPH(att.TimeOut.String)
	lunchbreakIn, lunchbreakOut := utils.ExtractTimeToPH(att.LunchBreakIn.String), utils.ExtractTimeToPH(att.LunchBreakOut.String)
	c := computeInOutStatus(timeIn, timeOut, schedIn, schedOut)
	lunchbreak := computeInOutStatus(lunchbreakIn, lunchbreakOut, "12:00", "13:00")

	var inShop, outShop bool
	if lat, lng, ok := utils.ParseLocation(att.InLocation); ok && s.shopLocation.RadiusMeters > 0 {
		inShop = utils.IsWithinRadius(lat, lng, s.shopLocation.Lat, s.shopLocation.Lng, s.shopLocation.RadiusMeters)
	}
	if lat, lng, ok := utils.ParseLocation(att.OutLocation); ok && s.shopLocation.RadiusMeters > 0 {
		outShop = utils.IsWithinRadius(lat, lng, s.shopLocation.Lat, s.shopLocation.Lng, s.shopLocation.RadiusMeters)
	}

	var lbInShop, lbOutShop bool
	if lat, lng, ok := utils.ParseLocation(att.LunchBreakInLocation); ok && s.shopLocation.RadiusMeters > 0 {
		lbInShop = utils.IsWithinRadius(lat, lng, s.shopLocation.Lat, s.shopLocation.Lng, s.shopLocation.RadiusMeters)
	}
	if lat, lng, ok := utils.ParseLocation(att.LunchBreakOutLocation); ok && s.shopLocation.RadiusMeters > 0 {
		lbOutShop = utils.IsWithinRadius(lat, lng, s.shopLocation.Lat, s.shopLocation.Lng, s.shopLocation.RadiusMeters)
	}

	var inDeviceInfo, outDeviceInfo, lbInDeviceInfo, lbOutDeviceInfo string
	if att.InBrowser.Valid {
		inDeviceInfo = utils.FormatUserAgentDevice(types.UserAgentInfo{
			Browser:        att.InBrowser.String,
			BrowserVersion: att.InBrowserVersion.String,
			OS:             att.InOs.String,
			Device:         att.InDevice.String,
		})
	}
	if att.OutBrowser.Valid {
		outDeviceInfo = utils.FormatUserAgentDevice(types.UserAgentInfo{
			Browser:        att.OutBrowser.String,
			BrowserVersion: att.OutBrowserVersion.String,
			OS:             att.OutOs.String,
			Device:         att.OutDevice.String,
		})
	}
	if att.LunchBreakInBrowser.Valid {
		lbInDeviceInfo = utils.FormatUserAgentDevice(types.UserAgentInfo{
			Browser:        att.LunchBreakInBrowser.String,
			BrowserVersion: att.LunchBreakInBrowserVersion.String,
			OS:             att.LunchBreakInOs.String,
			Device:         att.LunchBreakInDevice.String,
		})
	}
	if att.LunchBreakOutBrowser.Valid {
		lbOutDeviceInfo = utils.FormatUserAgentDevice(types.UserAgentInfo{
			Browser:        att.LunchBreakOutBrowser.String,
			BrowserVersion: att.LunchBreakOutBrowserVersion.String,
			OS:             att.LunchBreakOutOs.String,
			Device:         att.LunchBreakOutDevice.String,
		})
	}

	return models.Attendance{
		StaffID:          s.encoder.Encode(att.StaffID),
		FullName:         utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		Date:             att.ForDate,
		ScheduledTimeIn:  schedIn,
		ScheduledTimeOut: schedOut,

		Attendance: models.AttendanceStat{
			In:            timeIn,
			Out:           timeOut,
			InStatus:      c.inStatus,
			OutStatus:     c.outStatus,
			Duration:      c.duration,
			DurationColor: c.durationColor,
			InLate:        c.inLate,
			Undertime:     c.undertime,
			EarlyIn:       c.earlyIn,
			InShop:        inShop,
			OutShop:       outShop,
			InLocation:    att.InLocation.String,
			OutLocation:   att.OutLocation.String,
			InDeviceInfo:  inDeviceInfo,
			OutDeviceInfo: outDeviceInfo,
		},

		LunchBreak: models.AttendanceStat{
			In:            lunchbreakIn,
			Out:           lunchbreakOut,
			InStatus:      lunchbreak.inStatus,
			OutStatus:     lunchbreak.outStatus,
			Duration:      lunchbreak.duration,
			DurationColor: lunchbreak.durationColor,
			InShop:        lbInShop,
			OutShop:       lbOutShop,
			InLocation:    att.LunchBreakInLocation.String,
			OutLocation:   att.LunchBreakOutLocation.String,
			InDeviceInfo:  lbInDeviceInfo,
			OutDeviceInfo: lbOutDeviceInfo,
		},
	}
}

func computeInOutStatus(actualIn, actualOut, schedIn, schedOut string) attendanceStatusResult {
	out := attendanceStatusResult{duration: "-"}
	actualInM, inOk := utils.TimeToMinutes(actualIn)
	actualOutM, outOk := utils.TimeToMinutes(actualOut)
	schedInM, schedInOk := utils.TimeToMinutes(schedIn)
	schedOutM, schedOutOk := utils.TimeToMinutes(schedOut)

	if inOk && schedInOk {
		switch {
		case actualInM < schedInM:
			out.inStatus = enums.TIME_IN_STATUS_EARLIER
			timeSchedIn, err := time.Parse(constants.TimeLayoutHHMM, schedIn)
			if err != nil {
				logs.Log().Warn("computeInOutStatus", zap.String("sched in", schedIn), zap.Error(err))
			}
			timeActualIn, err := time.Parse(constants.TimeLayoutHHMMSS, actualIn)
			if err != nil {
				logs.Log().Warn("computeInOutStatus", zap.String("actual in", actualIn), zap.Error(err))
			}
			out.earlyIn = timeSchedIn.Sub(timeActualIn)
		case actualInM == schedInM:
			out.inStatus = enums.TIME_IN_STATUS_ON_TIME
		default:
			out.inStatus = enums.TIME_IN_STATUS_LATE

			timeSchedIn, err := time.Parse(constants.TimeLayoutHHMM, schedIn)
			if err != nil {
				logs.Log().Warn("computeInOutStatus", zap.String("sched in", schedIn), zap.Error(err))
			}
			timeActualIn, err := time.Parse(constants.TimeLayoutHHMMSS, actualIn)
			if err != nil {
				logs.Log().Warn("computeInOutStatus", zap.String("actual in", actualIn), zap.Error(err))
			}
			out.inLate = timeActualIn.Sub(timeSchedIn)
		}
	}

	if outOk && schedOutOk {
		switch {
		case actualOutM < schedOutM:
			out.outStatus = enums.TIME_OUT_STATUS_UNDERTIME
			timeSchedOut, err := time.Parse(constants.TimeLayoutHHMM, schedOut)
			if err != nil {
				logs.Log().Warn("computeInOutStatus", zap.String("sched out", schedOut), zap.Error(err))
			}
			timeActualOut, err := time.Parse(constants.TimeLayoutHHMMSS, actualOut)
			if err != nil {
				logs.Log().Warn("computeInOutStatus", zap.String("actual out", actualOut), zap.Error(err))
			}
			out.undertime = timeSchedOut.Sub(timeActualOut)
		case actualOutM == schedOutM:
			out.outStatus = enums.TIME_OUT_STATUS_ON_TIME
		default:
			out.outStatus = enums.TIME_OUT_STATUS_OVERTIME
		}
	}

	if inOk && outOk {
		diff := actualOutM - actualInM
		if diff >= 0 {
			out.duration = utils.FormatDurationFromMinutes(diff)
		}
		if schedInOk && schedOutOk {
			schedDuration := schedOutM - schedInM
			if schedDuration >= 0 && diff >= 0 {
				if diff >= schedDuration {
					out.durationColor = "green"
				} else {
					out.durationColor = "red"
				}
			}
		}
	}

	return out
}

type AttendanceExtraStats struct {
	TotalUndertimeMinutes float64
	TotalLateMinutes      float64
	TotalUndertimeCount   int
	TotalLateCount        int
	TotalEarlyInCount     int
	TotalOvertimeCount    int
}

func (s *AttendanceService) GetExtraStats(ctx context.Context, staffID string, data []staff.StaffRow) AttendanceExtraStats {
	var res AttendanceExtraStats

	decodedStaffID := s.encoder.Decode(staffID)
	staffDB, err := s.dbRO.GetQueries().GetStaffByID(ctx, decodedStaffID)
	if err != nil {
		return res
	}

	for _, d := range data {
		c := s.ComputeData(staff.StaffRowBase(staffDB), d)
		res.TotalLateMinutes += c.Attendance.InLate.Minutes()
		if c.Attendance.InStatus == enums.TIME_IN_STATUS_LATE {
			res.TotalLateCount++
		}
		if c.Attendance.OutStatus == enums.TIME_OUT_STATUS_UNDERTIME {
			res.TotalUndertimeCount++
			res.TotalUndertimeMinutes += c.Attendance.Undertime.Minutes()
		}
		if c.Attendance.OutStatus == enums.TIME_OUT_STATUS_OVERTIME {
			res.TotalOvertimeCount++
		}
		if c.Attendance.InStatus == enums.TIME_IN_STATUS_EARLIER {
			res.TotalEarlyInCount++
		}
	}
	return res
}
