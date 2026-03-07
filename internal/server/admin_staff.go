package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/services"
	"cchoice/internal/types"
	"cchoice/internal/utils"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

func buildStaffDayAttendance(
	encoder encode.IEncode,
	staff queries.GetStaffByIDRow,
	att queries.GetStaffAttendanceByDateRow,
) models.Attendance {
	schedIn, schedOut := "", ""
	if staff.TimeInSchedule.Valid {
		schedIn = staff.TimeInSchedule.String
	}
	if staff.TimeOutSchedule.Valid {
		schedOut = staff.TimeOutSchedule.String
	}
	timeIn, timeOut := utils.ExtractTimeToPH(att.TimeIn.String), utils.ExtractTimeToPH(att.TimeOut.String)
	c := computeAttendanceStatus(timeIn, timeOut, schedIn, schedOut)
	return models.Attendance{
		StaffID:          encoder.Encode(att.StaffID),
		FullName:         utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		Date:             att.ForDate,
		TimeIn:           timeIn,
		TimeOut:          timeOut,
		ScheduledTimeIn:  schedIn,
		ScheduledTimeOut: schedOut,
		TimeInStatus:     c.timeInStatus,
		TimeOutStatus:    c.timeOutStatus,
		Duration:         c.duration,
		DurationColor:    c.durationColor,
	}
}

func (s *Server) adminStaffHomeHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Home Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("staff_id", staffID), zap.Error(err))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}
	currentUserFullName := utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName)

	if err := compadmin.AdminStaffHomePage(currentUserFullName).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminStaffProfileHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Profile Handler]"
	ctx := r.Context()

	staff, staffID, err := s.getCurrentStaff(r.Context())
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.Int64("staff_id", staffID))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}

	profile := models.AdminStaffProfile{
		FullName:         utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		Birthdate:        staff.Birthdate,
		DateHired:        staff.DateHired,
		Position:         staff.Position,
		Email:            staff.Email,
		MobileNo:         staff.MobileNo,
		ScheduledTimeIn:  staff.TimeInSchedule.String,
		ScheduledTimeOut: staff.TimeOutSchedule.String,
		RequireInShop:    staff.RequireInShop,
		UserType:         enums.ParseStaffUserTypeToEnum(staff.UserType),
	}

	if err := compadmin.AdminHeader(&profile).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Page Handler]"
	ctx := r.Context()

	staff, staffID, err := s.getCurrentStaff(r.Context())
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.Int64("staff_id", staffID))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}

	today := time.Now().Format(constants.DateLayoutISO)
	attendance, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: today,
	})
	var hasTimeIn, hasTimeOut bool
	var myAttendance *models.Attendance
	if err != nil {
		if err != sql.ErrNoRows {
			logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
			http.Redirect(w, r, utils.URL("/admin/staff"), http.StatusSeeOther)
			return
		}
	} else {
		hasTimeIn = attendance.TimeIn.Valid
		hasTimeOut = attendance.TimeOut.Valid
		rec := buildStaffDayAttendance(s.encoder, staff, attendance)
		myAttendance = &rec
	}

	shop := conf.Conf().Settings.ShopLocation
	var inShop, outShop *bool
	if shop.RadiusMeters > 0 {
		if err == nil {
			if lat, lng, ok := utils.ParseLocation(attendance.InLocation); ok {
				b := utils.IsWithinRadius(lat, lng, shop.Lat, shop.Lng, shop.RadiusMeters)
				inShop = &b
			}
			if lat, lng, ok := utils.ParseLocation(attendance.OutLocation); ok {
				b := utils.IsWithinRadius(lat, lng, shop.Lat, shop.Lng, shop.RadiusMeters)
				outShop = &b
			}
		}
		if inShop == nil {
			locJSON := GetLocation(ctx, s.sessionManager)
			if lat, lng, ok := utils.ParseLocation(locJSON); ok {
				b := utils.IsWithinRadius(lat, lng, shop.Lat, shop.Lng, shop.RadiusMeters)
				inShop = &b
			}
		}
		if outShop == nil {
			locJSON := GetLocation(ctx, s.sessionManager)
			if lat, lng, ok := utils.ParseLocation(locJSON); ok {
				b := utils.IsWithinRadius(lat, lng, shop.Lat, shop.Lng, shop.RadiusMeters)
				outShop = &b
			}
		}
	}

	scheduledTimeIn := staff.TimeInSchedule.String
	scheduledTimeOut := staff.TimeOutSchedule.String

	locationDisplay := "unable to get location"
	distanceMeters := 0.0
	if locJSON := GetLocation(ctx, s.sessionManager); locJSON.Valid {
		if lat, lng, ok := utils.ParseLocation(locJSON); ok {
			locationDisplay = fmt.Sprintf("%.4f, %.4f", lat, lng)
			if shop.Lat != 0 && shop.Lng != 0 {
				distanceMeters = utils.HaversineDistanceMeters(lat, lng, shop.Lat, shop.Lng)
			}
		} else {
			locationDisplay = locJSON.String
		}
	}

	canTimeIn := !hasTimeIn
	canTimeOut := hasTimeIn && !hasTimeOut
	//INFO: (flam) - allow for now
	// if staff.RequireInShop {
	// 	if inShop == nil || !*inShop {
	// 		canTimeIn = false
	// 		canTimeOut = false
	// 	}
	// }

	profile := models.AdminStaffProfile{
		FullName:         utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		Birthdate:        staff.Birthdate,
		DateHired:        staff.DateHired,
		Position:         staff.Position,
		Email:            staff.Email,
		MobileNo:         staff.MobileNo,
		ScheduledTimeIn:  scheduledTimeIn,
		ScheduledTimeOut: scheduledTimeOut,
		SelectedDate:     today,
		CurrentDate:      time.Now().Format(constants.DateLayoutDisplay),
		CurrentTime:      time.Now().Format(constants.TimeLayoutDisplay),
		HasTimeIn:        hasTimeIn,
		HasTimeOut:       hasTimeOut,
		CanTimeIn:        canTimeIn,
		CanTimeOut:       canTimeOut,
		RequireInShop:    staff.RequireInShop,
		MyAttendance:     myAttendance,
		InShop:           inShop,
		OutShop:          outShop,
		LocationDisplay:  locationDisplay,
		DistanceMeters:   distanceMeters,
		UserType:         enums.ParseStaffUserTypeToEnum(staff.UserType),
	}

	if err := compadmin.AdminStaffPage(profile).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffAttendanceTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Attendance Handler]"
	ctx := r.Context()

	staff, staffID, err := s.getCurrentStaff(r.Context())
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.Int64("staff_id", staffID))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}

	date := utils.ParseAttendanceDate(r.URL.Query().Get("date"))
	attendance, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: date,
	})
	var record *models.Attendance
	if err == nil {
		rec := buildStaffDayAttendance(s.encoder, staff, attendance)
		record = &rec
	}

	if err := compadmin.StaffAttendanceSingleTable(record).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffTimeInHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	now := time.Now().UTC().Format(constants.DateTimeLayoutISO)
	date := utils.NowPH().Format(constants.DateLayoutISO)
	location := GetLocation(ctx, s.sessionManager)
	useragentID := getOrCreateUserAgentID(ctx, s.dbRW, r.UserAgent())
	svc := services.NewAttendanceService(s.dbRO, s.dbRW)
	err := svc.TimeIn(ctx, staffID, date, now, location, useragentID)
	if err != nil {
		http.Error(w, "Unable to time in", http.StatusBadRequest)
		return
	}
	w.Header().Set("HX-Redirect", utils.URL("/admin/staff/attendance"))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) adminStaffTimeOutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	now := time.Now().UTC().Format(constants.DateTimeLayoutISO)
	date := utils.NowPH().Format(constants.DateLayoutISO)
	useragentID := getOrCreateUserAgentID(ctx, s.dbRW, r.UserAgent())
	svc := services.NewAttendanceService(s.dbRO, s.dbRW)
	err := svc.TimeOut(ctx, staffID, date, now, useragentID)
	if err != nil {
		http.Error(w, "Unable to time out", http.StatusBadRequest)
		return
	}
	w.Header().Set("HX-Redirect", utils.URL("/admin/staff/attendance"))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) adminStaffTimeOffPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Request Time Off Page Handler]"
	ctx := r.Context()
	if err := compadmin.AdminStaffRequestTimeOffPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffTimeOffHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Time Off Handler]"
	ctx := r.Context()
	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	timeOffType := enums.ParseTimeOffToEnum(r.PostFormValue("type"))
	if timeOffType == enums.TIME_OFF_UNDEFINED {
		http.Error(w, "Invalid time off type", http.StatusBadRequest)
		return
	}
	description := r.PostFormValue("description")
	if description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}
	startDate, err := time.Parse(constants.DateLayoutISO, r.PostFormValue("start-date"))
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse(constants.DateLayoutISO, r.PostFormValue("end-date"))
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}
	if endDate.Before(startDate) {
		http.Error(w, "end date must not be before start date", http.StatusBadRequest)
		return
	}
	useragentID := getOrCreateUserAgentID(ctx, s.dbRW, r.UserAgent())
	svc := services.NewAttendanceService(s.dbRO, s.dbRW)
	if err := svc.TimeOff(
		ctx,
		staffID,
		timeOffType,
		description,
		startDate,
		endDate,
		useragentID,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Unable to time off", http.StatusBadRequest)
		return
	}
	w.Header().Set("HX-Redirect", utils.URL("/admin/staff/time-off"))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) adminStaffTimeOffTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Time Off Table Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	if staffID == 0 {
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}

	timeOffs, err := s.dbRO.GetQueries().GetStaffTimeOffsByStaffID(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.Int64("staff_id", staffID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	staffTimeOffs := make([]models.StaffTimeOff, 0, len(timeOffs))
	for _, to := range timeOffs {
		var approvedBy string
		var approvedAt string

		if to.ApprovedBy.Valid && to.ApproverFirstName.Valid {
			approvedBy = utils.BuildFullName(
				to.ApproverFirstName.String,
				to.ApproverMiddleName.String,
				to.ApproverLastName.String,
			)
		} else {
			approvedBy = "-"
		}

		if to.ApprovedAt.Valid {
			approvedAt = to.ApprovedAt.Time.Format(constants.DateTimeLayoutISO)
		} else {
			approvedAt = "-"
		}

		staffTimeOffs = append(staffTimeOffs, models.StaffTimeOff{
			ID:          s.encoder.Encode(to.ID),
			Type:        enums.ParseTimeOffToEnum(to.Type),
			CreatedAt:   utils.ConvertToPH(to.CreatedAt),
			StartDate:   to.StartDate.Format(constants.DateLayoutISO),
			EndDate:     to.EndDate.Format(constants.DateLayoutISO),
			Description: to.Description,
			Approved:    to.Approved.Bool,
			ApprovedBy:  approvedBy,
			ApprovedAt:  approvedAt,
		})
	}

	if err := compadmin.StaffTimeOffsTable(staffTimeOffs).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffAttendanceLocationHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Attendance Location Handler]"

	ctx := r.Context()
	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	if staffID == 0 {
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}

	var date, latStr, lngStr string
	if r.Header.Get("Content-Type") == "application/json" {
		var body struct {
			Date string  `json:"date"`
			Lat  float64 `json:"lat"`
			Lng  float64 `json:"lng"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}
		date = body.Date
		latStr = strconv.FormatFloat(body.Lat, 'f', -1, 64)
		lngStr = strconv.FormatFloat(body.Lng, 'f', -1, 64)
	} else {
		_ = r.ParseForm()
		date = r.PostFormValue("date")
		latStr = r.PostFormValue("lat")
		lngStr = r.PostFormValue("lng")
	}

	date = utils.ParseAttendanceDate(date)

	if latStr == "" || lngStr == "" {
		http.Error(w, "lat and lng required", http.StatusBadRequest)
		return
	}

	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "invalid lat/lng", http.StatusBadRequest)
		return
	}

	locationJSON, _ := json.Marshal(types.Location{Lat: lat, Lng: lng})
	location := sql.NullString{String: string(locationJSON), Valid: true}

	attendanceService := services.NewAttendanceService(s.dbRO, s.dbRW)
	_ = attendanceService.UpsertLocation(ctx, staffID, date, location)

	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.Int64("staff_id", staffID))
		http.Error(w, "failed to get staff", http.StatusBadRequest)
		return
	}

	locationResult := services.ComputeLocation(
		location,
		GetLocation(ctx, s.sessionManager),
		conf.Conf().Settings.ShopLocation,
	)

	profile := models.AdminStaffProfile{
		FullName:        utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		RequireInShop:   staff.RequireInShop,
		InShop:          locationResult.InShop,
		LocationDisplay: locationResult.LocationDisplay,
		DistanceMeters:  locationResult.DistanceMeters,
		UserType:        enums.ParseStaffUserTypeToEnum(staff.UserType),
	}

	_ = compadmin.StaffLocationStatus(profile).Render(ctx, w)
}
