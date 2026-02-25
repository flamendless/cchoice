package server

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"cchoice/cmd/web/components"
	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

const (
	SessionStaffID     = "staff_id"
	SessionLocationLat = "location_lat"
	SessionLocationLng = "location_lng"
	maxStaffListSize   = 1000
)

type attendanceStatusResult struct {
	timeInStatus  enums.TimeInStatus
	timeOutStatus enums.TimeOutStatus
	duration      string
	durationColor string
}

func computeAttendanceStatus(actualIn, actualOut, schedIn, schedOut string) attendanceStatusResult {
	out := attendanceStatusResult{duration: "-"}
	actualInM, inOk := utils.TimeToMinutes(actualIn)
	actualOutM, outOk := utils.TimeToMinutes(actualOut)
	schedInM, schedInOk := utils.TimeToMinutes(schedIn)
	schedOutM, schedOutOk := utils.TimeToMinutes(schedOut)

	if inOk && schedInOk {
		switch {
		case actualInM < schedInM:
			out.timeInStatus = enums.TIME_IN_STATUS_EARLIER
		case actualInM == schedInM:
			out.timeInStatus = enums.TIME_IN_STATUS_ON_TIME
		default:
			out.timeInStatus = enums.TIME_IN_STATUS_LATE
		}
	}
	if outOk && schedOutOk {
		switch {
		case actualOutM < schedOutM:
			out.timeOutStatus = enums.TIME_OUT_STATUS_UNDERTIME
		case actualOutM == schedOutM:
			out.timeOutStatus = enums.TIME_OUT_STATUS_ON_TIME
		default:
			out.timeOutStatus = enums.TIME_OUT_STATUS_OVERTIME
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

func parseAttendanceDate(date string) string {
	if date == "" {
		return time.Now().Format(constants.DateLayoutISO)
	}
	if _, err := time.Parse(constants.DateLayoutISO, date); err != nil {
		return time.Now().Format(constants.DateLayoutISO)
	}
	return date
}

type attendanceLocationJSON struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func parseAttendanceLocation(locationJSON sql.NullString) (lat, lng float64, ok bool) {
	if !locationJSON.Valid || locationJSON.String == "" {
		return 0, 0, false
	}
	var loc attendanceLocationJSON
	if err := json.Unmarshal([]byte(locationJSON.String), &loc); err != nil {
		return 0, 0, false
	}
	return loc.Lat, loc.Lng, true
}

func sessionLocationJSON(ctx context.Context, get func(context.Context, string) interface{}) sql.NullString {
	latVal := get(ctx, SessionLocationLat)
	lngVal := get(ctx, SessionLocationLng)
	if latVal == nil || lngVal == nil {
		return sql.NullString{}
	}
	lat, ok1 := latVal.(float64)
	lng, ok2 := lngVal.(float64)
	if !ok1 || !ok2 {
		return sql.NullString{}
	}
	b, _ := json.Marshal(attendanceLocationJSON{Lat: lat, Lng: lng})
	return sql.NullString{String: string(b), Valid: true}
}

func buildAdminStaffAttendance(staff queries.GetAllStaffsRow, att queries.GetStaffAttendanceByStaffIDAndDateRangeRow, shop conf.ShopLocation) models.AdminStaffAttendance {
	schedIn, schedOut := "", ""
	if staff.TimeInSchedule.Valid {
		schedIn = staff.TimeInSchedule.String
	}
	if staff.TimeOutSchedule.Valid {
		schedOut = staff.TimeOutSchedule.String
	}
	c := computeAttendanceStatus(att.TimeIn.String, att.TimeOut.String, schedIn, schedOut)
	inShop := false
	if lat, lng, ok := parseAttendanceLocation(att.Location); ok && shop.RadiusMeters > 0 {
		inShop = utils.IsWithinRadius(lat, lng, shop.Lat, shop.Lng, shop.RadiusMeters)
	}
	return models.AdminStaffAttendance{
		StaffID:          att.StaffID,
		FullName:         utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		TimeIn:           att.TimeIn.String,
		TimeOut:          att.TimeOut.String,
		ScheduledTimeIn:  schedIn,
		ScheduledTimeOut: schedOut,
		TimeInStatus:     c.timeInStatus,
		TimeOutStatus:    c.timeOutStatus,
		Duration:         c.duration,
		DurationColor:    c.durationColor,
		InShop:           inShop,
	}
}

func buildStaffDayAttendance(staff queries.GetStaffByIDRow, att queries.GetStaffAttendanceByDateRow) models.AdminStaffAttendance {
	schedIn, schedOut := "", ""
	if staff.TimeInSchedule.Valid {
		schedIn = staff.TimeInSchedule.String
	}
	if staff.TimeOutSchedule.Valid {
		schedOut = staff.TimeOutSchedule.String
	}
	c := computeAttendanceStatus(att.TimeIn.String, att.TimeOut.String, schedIn, schedOut)
	return models.AdminStaffAttendance{
		StaffID:          staff.ID,
		FullName:         utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		TimeIn:           att.TimeIn.String,
		TimeOut:          att.TimeOut.String,
		ScheduledTimeIn:  schedIn,
		ScheduledTimeOut: schedOut,
		TimeInStatus:     c.timeInStatus,
		TimeOutStatus:    c.timeOutStatus,
		Duration:         c.duration,
		DurationColor:    c.durationColor,
	}
}

func AddAdminHandlers(s *Server, r chi.Router) {
	r.Get("/admin", s.adminLoginPageHandler)
	r.Post("/admin/login", s.adminLoginHandler)
	r.With(s.requireStaffAuth).Post("/admin/logout", s.adminLogoutHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff", s.adminStaffPageHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff/attendance", s.adminStaffAttendanceHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/time-in", s.adminStaffTimeInHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/time-out", s.adminStaffTimeOutHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/attendance/location", s.adminStaffAttendanceLocationHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser", s.adminSuperuserPageHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/attendance", s.adminSuperuserAttendanceHandler)
}

func (s *Server) requireStaffAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		staffID := s.sessionManager.GetInt64(r.Context(), SessionStaffID)
		if staffID == 0 {
			http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireSuperuserAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		staffID := s.sessionManager.GetInt64(r.Context(), SessionStaffID)
		if staffID == 0 {
			http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
			return
		}

		staff, err := s.dbRO.GetQueries().GetStaffByID(r.Context(), staffID)
		if err != nil || staff.UserType != enums.STAFF_USER_TYPE_SUPERUSER.String() {
			http.Redirect(w, r, utils.URL("/admin/staff"), http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) adminLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Login Page Handler]"
	ctx := r.Context()

	loginError := ""
	switch r.URL.Query().Get("error") {
	case "invalid_credentials":
		loginError = "Invalid email or password."
	case "invalid_format":
		loginError = "Invalid email or password format."
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.AdminLoginPage(loginError).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Login Handler]"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Redirect(w, r, utils.URL("/admin?error=invalid_format"), http.StatusSeeOther)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	if !constants.EmailRegex.MatchString(email) {
		http.Redirect(w, r, utils.URL("/admin?error=invalid_format"), http.StatusSeeOther)
		return
	}

	if !constants.PasswordRegex.MatchString(password) {
		http.Redirect(w, r, utils.URL("/admin?error=invalid_format"), http.StatusSeeOther)
		return
	}

	staff, err := s.dbRO.GetQueries().GetStaffByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Redirect(w, r, utils.URL("/admin?error=invalid_credentials"), http.StatusSeeOther)
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.Password), []byte(password)); err != nil {
		http.Redirect(w, r, utils.URL("/admin?error=invalid_credentials"), http.StatusSeeOther)
		return
	}

	s.sessionManager.Put(ctx, SessionStaffID, staff.ID)

	if latStr, lngStr := r.PostFormValue("location_lat"), r.PostFormValue("location_lng"); latStr != "" && lngStr != "" {
		if lat, errLat := strconv.ParseFloat(latStr, 64); errLat == nil {
			if lng, errLng := strconv.ParseFloat(lngStr, 64); errLng == nil {
				s.sessionManager.Put(ctx, SessionLocationLat, lat)
				s.sessionManager.Put(ctx, SessionLocationLng, lng)
			}
		}
	}

	switch enums.ParseStaffUserTypeToEnum(staff.UserType) {
	case enums.STAFF_USER_TYPE_SUPERUSER:
		http.Redirect(w, r, utils.URL("/admin/superuser"), http.StatusSeeOther)
	case enums.STAFF_USER_TYPE_STAFF:
		http.Redirect(w, r, utils.URL("/admin/staff"), http.StatusSeeOther)
	default:
		logs.Log().Warn(logtag, zap.String("got unhandled", staff.UserType))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
	}
}

func (s *Server) adminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Logout Handler]"
	ctx := r.Context()

	if err := s.sessionManager.Destroy(ctx); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}

	http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
}

func (s *Server) adminStaffPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Page Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}

	today := time.Now().Format(constants.DateLayoutISO)
	attendance, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: today,
	})
	var hasTimeIn, hasTimeOut bool
	var myAttendance *models.AdminStaffAttendance
	if err != nil {
		if err != sql.ErrNoRows {
			logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
			http.Redirect(w, r, utils.URL("/admin/staff"), http.StatusSeeOther)
			return
		}
	} else {
		hasTimeIn = attendance.TimeIn.Valid
		hasTimeOut = attendance.TimeOut.Valid
		rec := buildStaffDayAttendance(staff, attendance)
		myAttendance = &rec
	}

	shop := conf.Conf().Settings.ShopLocation
	var inShop *bool
	if shop.RadiusMeters > 0 {
		if err == nil {
			if lat, lng, ok := parseAttendanceLocation(attendance.Location); ok {
				b := utils.IsWithinRadius(lat, lng, shop.Lat, shop.Lng, shop.RadiusMeters)
				inShop = &b
			}
		}
		if inShop == nil {
			locJSON := sessionLocationJSON(ctx, func(c context.Context, k string) interface{} { return s.sessionManager.Get(c, k) })
			if lat, lng, ok := parseAttendanceLocation(locJSON); ok {
				b := utils.IsWithinRadius(lat, lng, shop.Lat, shop.Lng, shop.RadiusMeters)
				inShop = &b
			}
		}
	}

	scheduledTimeIn := ""
	if staff.TimeInSchedule.Valid {
		scheduledTimeIn = staff.TimeInSchedule.String
	}
	scheduledTimeOut := ""
	if staff.TimeOutSchedule.Valid {
		scheduledTimeOut = staff.TimeOutSchedule.String
	}

	profile := models.AdminStaffProfile{
		FullName:         utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		Birthdate:        staff.Birthdate,
		DateHired:        staff.DateHired,
		Position:         staff.Position,
		ScheduledTimeIn:  scheduledTimeIn,
		ScheduledTimeOut: scheduledTimeOut,
		SelectedDate:     today,
		CurrentDate:      time.Now().Format(constants.DateLayoutDisplay),
		CurrentTime:      time.Now().Format(constants.TimeLayoutDisplay),
		HasTimeIn:        hasTimeIn,
		HasTimeOut:       hasTimeOut,
		CanTimeIn:        !hasTimeIn,
		CanTimeOut:       hasTimeIn && !hasTimeOut,
		MyAttendance:     myAttendance,
		InShop:           inShop,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.AdminStaffPage(profile).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Attendance Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}

	date := parseAttendanceDate(r.URL.Query().Get("date"))

	attendance, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: date,
	})
	var record *models.AdminStaffAttendance
	if err == nil {
		rec := buildStaffDayAttendance(staff, attendance)
		record = &rec
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.StaffAttendanceSingleTable(record).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffTimeInHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Time In Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	today := time.Now().Format(constants.DateLayoutISO)
	now := time.Now().Format(constants.DateTimeLayoutISO)

	existing, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: today,
	})
	if err != nil && err != sql.ErrNoRows {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	locationFromSession := sessionLocationJSON(ctx, func(c context.Context, k string) interface{} { return s.sessionManager.Get(c, k) })

	if err == nil && existing.TimeIn.Valid {
		http.Error(w, "Time in already recorded for today", http.StatusBadRequest)
		return
	}

	if err == sql.ErrNoRows {
		_, err = s.dbRW.GetQueries().CreateStaffAttendance(ctx, queries.CreateStaffAttendanceParams{
			StaffID:  staffID,
			ForDate:  today,
			TimeIn:   sql.NullString{String: now, Valid: true},
			TimeOut:  sql.NullString{},
			Location: locationFromSession,
		})
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		_, err = s.dbRW.GetQueries().UpdateStaffAttendanceTimeIn(ctx, queries.UpdateStaffAttendanceTimeInParams{
			TimeIn:   sql.NullString{String: now, Valid: true},
			Location: locationFromSession,
			StaffID:  staffID,
			ForDate:  today,
		})
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", utils.URL("/admin/staff"))
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, utils.URL("/admin/staff"), http.StatusSeeOther)
}

func (s *Server) adminStaffTimeOutHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Time Out Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	today := time.Now().Format(constants.DateLayoutISO)
	now := time.Now().Format(constants.DateTimeLayoutISO)

	existing, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: today,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Time in required before time out", http.StatusBadRequest)
		return
	}

	if !existing.TimeIn.Valid {
		http.Error(w, "Time in required before time out", http.StatusBadRequest)
		return
	}

	if existing.TimeOut.Valid {
		http.Error(w, "Time out already recorded for today", http.StatusBadRequest)
		return
	}

	_, err = s.dbRW.GetQueries().UpdateStaffAttendanceTimeOut(ctx, queries.UpdateStaffAttendanceTimeOutParams{
		TimeOut: sql.NullString{String: now, Valid: true},
		StaffID: staffID,
		ForDate: today,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", utils.URL("/admin/staff"))
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, utils.URL("/admin/staff"), http.StatusSeeOther)
}

func (s *Server) adminSuperuserPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Page Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	currentStaff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("staff_id", staffID), zap.Error(err))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}
	currentUserFullName := utils.BuildFullName(currentStaff.FirstName, currentStaff.MiddleName.String, currentStaff.LastName)

	today := time.Now().Format(constants.DateLayoutISO)

	attendances, err := s.dbRO.GetQueries().GetStaffAttendanceByStaffIDAndDateRange(ctx, today)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		attendances = []queries.GetStaffAttendanceByStaffIDAndDateRangeRow{}
	}

	staffs, err := s.dbRO.GetQueries().GetAllStaffs(ctx, maxStaffListSize)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		staffs = []queries.GetAllStaffsRow{}
	}

	staffMap := make(map[int64]queries.GetAllStaffsRow)
	for _, staff := range staffs {
		staffMap[staff.ID] = staff
	}

	shop := conf.Conf().Settings.ShopLocation
	attendanceData := make([]models.AdminStaffAttendance, 0, len(attendances))
	for _, att := range attendances {
		staff, ok := staffMap[att.StaffID]
		if !ok {
			continue
		}
		attendanceData = append(attendanceData, buildAdminStaffAttendance(staff, att, shop))
	}

	pageData := models.AdminSuperuserPage{
		FullName:     currentUserFullName,
		CurrentDate:  time.Now().Format(constants.DateLayoutDisplay),
		CurrentTime:  time.Now().Format(constants.TimeLayoutDisplay),
		SelectedDate: today,
		Attendances:  attendanceData,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.AdminSuperuserPage(pageData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminSuperuserAttendanceHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Attendance Handler]"
	ctx := r.Context()

	date := parseAttendanceDate(r.URL.Query().Get("date"))

	attendances, err := s.dbRO.GetQueries().GetStaffAttendanceByStaffIDAndDateRange(ctx, date)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		attendances = []queries.GetStaffAttendanceByStaffIDAndDateRangeRow{}
	}

	staffs, err := s.dbRO.GetQueries().GetAllStaffs(ctx, maxStaffListSize)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		staffs = []queries.GetAllStaffsRow{}
	}

	staffMap := make(map[int64]queries.GetAllStaffsRow)
	for _, staff := range staffs {
		staffMap[staff.ID] = staff
	}

	shop := conf.Conf().Settings.ShopLocation
	attendanceData := make([]models.AdminStaffAttendance, 0, len(attendances))
	for _, att := range attendances {
		staff, ok := staffMap[att.StaffID]
		if !ok {
			continue
		}
		attendanceData = append(attendanceData, buildAdminStaffAttendance(staff, att, shop))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.AdminSuperuserAttendanceTable(attendanceData).Render(ctx, w); err != nil {
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
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		date, latStr, lngStr = body.Date, strconv.FormatFloat(body.Lat, 'f', -1, 64), strconv.FormatFloat(body.Lng, 'f', -1, 64)
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form", http.StatusBadRequest)
			return
		}
		date = r.PostFormValue("date")
		latStr = r.PostFormValue("lat")
		lngStr = r.PostFormValue("lng")
	}

	date = parseAttendanceDate(date)
	if latStr == "" || lngStr == "" {
		http.Error(w, "lat and lng required", http.StatusBadRequest)
		return
	}
	lat, errLat := strconv.ParseFloat(latStr, 64)
	lng, errLng := strconv.ParseFloat(lngStr, 64)
	if errLat != nil || errLng != nil {
		http.Error(w, "Invalid lat or lng", http.StatusBadRequest)
		return
	}

	locJSON, _ := json.Marshal(attendanceLocationJSON{Lat: lat, Lng: lng})
	location := sql.NullString{String: string(locJSON), Valid: true}

	_, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: date,
	})
	if err != nil && err != sql.ErrNoRows {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err == sql.ErrNoRows {
		_, err = s.dbRW.GetQueries().CreateStaffAttendance(ctx, queries.CreateStaffAttendanceParams{
			StaffID:  staffID,
			ForDate:  date,
			TimeIn:   sql.NullString{},
			TimeOut:  sql.NullString{},
			Location: location,
		})
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		_, err = s.dbRW.GetQueries().UpdateStaffAttendanceLocation(ctx, queries.UpdateStaffAttendanceLocationParams{
			Location: location,
			StaffID:  staffID,
			ForDate:  date,
		})
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
