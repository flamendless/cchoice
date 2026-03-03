package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/types"
	"cchoice/internal/utils"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/chi/v5"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

const (
	SessionStaffID       = "staff_id"
	SessionStaffAccessID = "staff_access_id"
	maxStaffListSize     = 1000
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

func getOrCreateUserAgentID(ctx context.Context, db database.Service, userAgentStr string) sql.NullInt64 {
	if userAgentStr == "" {
		return sql.NullInt64{}
	}

	uaInfo := utils.ParseUserAgent(userAgentStr)
	if uaInfo.Browser == "" {
		return sql.NullInt64{}
	}

	id, err := db.GetQueries().UpsertUserAgent(ctx, queries.UpsertUserAgentParams{
		UserAgent:      userAgentStr,
		Browser:        uaInfo.Browser,
		BrowserVersion: uaInfo.BrowserVersion,
		Os:             uaInfo.OS,
		Device:         uaInfo.Device,
	})
	if err != nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: id, Valid: true}
}

func buildAdminStaffAttendance(
	staff queries.GetAllStaffsRow,
	att queries.GetStaffAttendanceByStaffIDAndDateRangeRow,
	shopLocation types.Location,
) models.Attendance {
	schedIn, schedOut := "", ""
	if staff.TimeInSchedule.Valid {
		schedIn = staff.TimeInSchedule.String
	}
	if staff.TimeOutSchedule.Valid {
		schedOut = staff.TimeOutSchedule.String
	}
	timeIn, timeOut := utils.ExtractTime(att.TimeIn.String), utils.ExtractTime(att.TimeOut.String)
	c := computeAttendanceStatus(timeIn, timeOut, schedIn, schedOut)
	inShop := false
	if lat, lng, ok := utils.ParseLocation(att.Location); ok && shopLocation.RadiusMeters > 0 {
		inShop = utils.IsWithinRadius(lat, lng, shopLocation.Lat, shopLocation.Lng, shopLocation.RadiusMeters)
	}

	deviceInfo := ""
	if att.Browser.Valid {
		deviceInfo = utils.FormatUserAgentDevice(types.UserAgentInfo{
			Browser:        att.Browser.String,
			BrowserVersion: att.BrowserVersion.String,
			OS:             att.Os.String,
			Device:         att.Device.String,
		})
	}

	return models.Attendance{
		StaffID:          att.StaffID,
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
		InShop:           inShop,
		Location:         att.Location.String,
		DeviceInfo:       deviceInfo,
	}
}

func buildStaffDayAttendance(staff queries.GetStaffByIDRow, att queries.GetStaffAttendanceByDateRow) models.Attendance {
	schedIn, schedOut := "", ""
	if staff.TimeInSchedule.Valid {
		schedIn = staff.TimeInSchedule.String
	}
	if staff.TimeOutSchedule.Valid {
		schedOut = staff.TimeOutSchedule.String
	}
	timeIn, timeOut := utils.ExtractTime(att.TimeIn.String), utils.ExtractTime(att.TimeOut.String)
	c := computeAttendanceStatus(timeIn, timeOut, schedIn, schedOut)
	return models.Attendance{
		StaffID:          staff.ID,
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

func AddAdminHandlers(s *Server, r chi.Router) {
	r.Get("/admin", s.adminLoginPageHandler)
	r.Post("/admin/login", s.adminLoginHandler)
	r.With(s.requireStaffAuth).Post("/admin/logout", s.adminLogoutHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff", s.adminStaffPageHandler)
	r.With(s.requireStaffAuth).Get("/admin/profile", s.adminStaffProfileHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff/attendance", s.adminStaffAttendanceHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/time-in", s.adminStaffTimeInHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/time-out", s.adminStaffTimeOutHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/attendance/location", s.adminStaffAttendanceLocationHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser", s.adminSuperuserHomeHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/attendance", s.adminSuperuserAttendancePageHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/attendance/table", s.adminSuperuserAttendanceHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/products", s.adminSuperuserProductsHandler)
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

	if err := compadmin.AdminLoginPage(loginError).Render(ctx, w); err != nil {
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
	htmx := r.Header.Get("HX-Request") == "true"
	redirect := func(url string) {
		if htmx {
			w.Header().Set("HX-Redirect", url)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, url, http.StatusSeeOther)
	}

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirect(utils.URL("/admin?error=invalid_format"))
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	if !constants.EmailRegex.MatchString(email) {
		redirect(utils.URL("/admin?error=invalid_format"))
		return
	}

	if !constants.PasswordRegex.MatchString(password) {
		redirect(utils.URL("/admin?error=invalid_format"))
		return
	}

	staff, err := s.dbRO.GetQueries().GetStaffByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			redirect(utils.URL("/admin?error=invalid_credentials"))
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.Password), []byte(password)); err != nil {
		redirect(utils.URL("/admin?error=invalid_credentials"))
		return
	}

	s.sessionManager.Put(ctx, SessionStaffID, staff.ID)

	useragentID := sql.NullInt64{}
	if ua := r.UserAgent(); ua != "" {
		useragentID = getOrCreateUserAgentID(context.Background(), s.dbRW, ua)
	}
	accessID, err := s.dbRW.GetQueries().CreateStaffAccess(context.Background(), queries.CreateStaffAccessParams{
		StaffID:     staff.ID,
		UseragentID: useragentID,
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	} else {
		s.sessionManager.Put(ctx, SessionStaffAccessID, accessID)
	}

	if latStr, lngStr := r.PostFormValue("location_lat"), r.PostFormValue("location_lng"); latStr != "" && lngStr != "" {
		SetLocation(ctx, s.sessionManager, latStr, lngStr)
	}

	switch enums.ParseStaffUserTypeToEnum(staff.UserType) {
	case enums.STAFF_USER_TYPE_SUPERUSER:
		redirect(utils.URL("/admin/superuser"))
	case enums.STAFF_USER_TYPE_STAFF:
		redirect(utils.URL("/admin/staff"))
	default:
		logs.Log().Warn(logtag, zap.String("got unhandled", staff.UserType))
		redirect(utils.URL("/admin"))
	}
}

func (s *Server) adminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Logout Handler]"
	ctx := r.Context()

	accessID := s.sessionManager.GetInt64(ctx, SessionStaffAccessID)
	if accessID != 0 {
		_, err := s.dbRW.GetQueries().UpdateStaffAccessLogout(ctx, accessID)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Int64("staff_access_id", accessID), zap.Error(err))
		}
	}

	if err := s.sessionManager.Destroy(ctx); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}

	http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
}

func (s *Server) adminStaffProfileHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Profile Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
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
		rec := buildStaffDayAttendance(staff, attendance)
		myAttendance = &rec
	}

	shop := conf.Conf().Settings.ShopLocation
	var inShop *bool
	if shop.RadiusMeters > 0 {
		if err == nil {
			if lat, lng, ok := utils.ParseLocation(attendance.Location); ok {
				b := utils.IsWithinRadius(lat, lng, shop.Lat, shop.Lng, shop.RadiusMeters)
				inShop = &b
			}
		}
		if inShop == nil {
			locJSON := GetLocation(ctx, s.sessionManager)
			if lat, lng, ok := utils.ParseLocation(locJSON); ok {
				b := utils.IsWithinRadius(lat, lng, shop.Lat, shop.Lng, shop.RadiusMeters)
				inShop = &b
			}
		}
	}

	scheduledTimeIn := staff.TimeInSchedule.String
	scheduledTimeOut := staff.TimeOutSchedule.String

	locationDisplay := ""
	if locJSON := GetLocation(ctx, s.sessionManager); locJSON.Valid {
		if lat, lng, ok := utils.ParseLocation(locJSON); ok {
			locationDisplay = fmt.Sprintf("%.4f, %.4f", lat, lng)
		}
	}

	canTimeIn := !hasTimeIn
	canTimeOut := hasTimeIn && !hasTimeOut
	if staff.RequireInShop {
		if inShop == nil || !*inShop {
			canTimeIn = false
			canTimeOut = false
		}
	}

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
		LocationDisplay:  locationDisplay,
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

	date := utils.ParseAttendanceDate(r.URL.Query().Get("date"))

	attendance, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: date,
	})
	var record *models.Attendance
	if err == nil {
		rec := buildStaffDayAttendance(staff, attendance)
		record = &rec
	}

	if err := compadmin.StaffAttendanceSingleTable(record).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffTimeInHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Time In Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	today := utils.NowPH().Format(constants.DateLayoutISO)
	now := utils.NowPH().Format(constants.DateTimeLayoutISO)

	existing, err := s.dbRO.GetQueries().GetStaffAttendanceByDate(ctx, queries.GetStaffAttendanceByDateParams{
		StaffID: staffID,
		ForDate: today,
	})
	if err != nil && err != sql.ErrNoRows {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	locationFromSession := GetLocation(ctx, s.sessionManager)

	userAgentStr := r.UserAgent()
	useragentID := getOrCreateUserAgentID(ctx, s.dbRW, userAgentStr)

	if err == nil && existing.TimeIn.Valid {
		http.Error(w, "Time in already recorded for today", http.StatusBadRequest)
		return
	}

	if err == sql.ErrNoRows {
		_, err = s.dbRW.GetQueries().CreateStaffAttendance(ctx, queries.CreateStaffAttendanceParams{
			StaffID:     staffID,
			ForDate:     today,
			TimeIn:      sql.NullString{String: now, Valid: true},
			TimeOut:     sql.NullString{},
			Location:    locationFromSession,
			UseragentID: useragentID,
		})
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		_, err = s.dbRW.GetQueries().UpdateStaffAttendanceTimeIn(ctx, queries.UpdateStaffAttendanceTimeInParams{
			TimeIn:      sql.NullString{String: now, Valid: true},
			Location:    locationFromSession,
			UseragentID: useragentID,
			StaffID:     staffID,
			ForDate:     today,
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
	today := utils.NowPH().Format(constants.DateLayoutISO)
	now := utils.NowPH().Format(constants.DateTimeLayoutISO)

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

	userAgentStr := r.UserAgent()
	useragentID := getOrCreateUserAgentID(ctx, s.dbRW, userAgentStr)

	_, err = s.dbRW.GetQueries().UpdateStaffAttendanceTimeOut(ctx, queries.UpdateStaffAttendanceTimeOutParams{
		TimeOut:     sql.NullString{String: now, Valid: true},
		UseragentID: useragentID,
		StaffID:     staffID,
		ForDate:     today,
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

func (s *Server) adminSuperuserHomeHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Home Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	currentStaff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("staff_id", staffID), zap.Error(err))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}
	currentUserFullName := utils.BuildFullName(currentStaff.FirstName, currentStaff.MiddleName.String, currentStaff.LastName)

	if err := compadmin.AdminSuperuserHomePage(currentUserFullName).Render(ctx, w); err != nil {
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

	date := utils.ParseAttendanceDate(r.URL.Query().Get("date"))

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
	attendanceData := make([]models.Attendance, 0, len(attendances))
	for _, att := range attendances {
		staff, ok := staffMap[att.StaffID]
		if !ok {
			continue
		}
		attendanceData = append(attendanceData, buildAdminStaffAttendance(staff, att, shop))
	}

	if err := compadmin.AdminSuperuserAttendanceTable(attendanceData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserAttendancePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Attendance Page Handler]"
	ctx := r.Context()
	date := utils.ParseAttendanceDate(r.URL.Query().Get("date"))

	if err := compadmin.AdminSuperuserAttendancePage("Staff Attendance", date).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserProductsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Handler]"
	ctx := r.Context()

	brandsRes, err := requests.GetBrandsForAdmin(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		requests.GenerateAdminBrandsCacheKey(),
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		brandsRes = []queries.GetBrandsForSidePanelRow{}
	}

	brands := make([]models.AdminBrand, 0, len(brandsRes))
	for _, b := range brandsRes {
		brands = append(brands, models.AdminBrand{
			ID:   b.ID,
			Name: b.Name,
		})
	}

	categoriesRes, err := requests.GetCategoriesForAdmin(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		requests.GenerateAdminCategoriesCacheKey(),
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		categoriesRes = map[string][]string{}
	}

	categories := make([]models.AdminCategory, 0, len(categoriesRes))
	for cat, subcats := range categoriesRes {
		categories = append(categories, models.AdminCategory{
			Category:      cat,
			Subcategories: subcats,
		})
	}

	formData := models.AdminProductForm{
		Brands:        brands,
		Categories:    categories,
		FormAction:    utils.URL("/admin/superuser/products"),
		CancelURL:     utils.URL("/admin/superuser"),
		VATPercentage: conf.Conf().Settings.VATPercentage,
	}

	if err := compadmin.AdminSuperuserProductsPage(formData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	date = utils.ParseAttendanceDate(date)
	if latStr == "" || lngStr == "" {
		staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
		if err != nil || staff.UserType != enums.STAFF_USER_TYPE_SUPERUSER.String() {
			http.Error(w, "lat and lng required", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	lat, errLat := strconv.ParseFloat(latStr, 64)
	lng, errLng := strconv.ParseFloat(lngStr, 64)
	if errLat != nil || errLng != nil {
		http.Error(w, "Invalid lat or lng", http.StatusBadRequest)
		return
	}

	locJSON, _ := json.Marshal(types.Location{Lat: lat, Lng: lng})
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
