package server

import (
	"context"
	"database/sql"
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/types"
	"cchoice/internal/utils"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/chi/v5"
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

func (s *Server) getCurrentStaff(ctx context.Context) (queries.GetStaffByIDRow, int64, error) {
	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
	return staff, staffID, err
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
	encoder encode.IEncode,
	staff queries.GetAllStaffsRow,
	att queries.GetStaffAttendanceByDateRangeRow,
	shopLocation types.Location,
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
	if lat, lng, ok := utils.ParseLocation(att.InLocation); ok && shopLocation.RadiusMeters > 0 {
		inShop = utils.IsWithinRadius(lat, lng, shopLocation.Lat, shopLocation.Lng, shopLocation.RadiusMeters)
	}
	if lat, lng, ok := utils.ParseLocation(att.OutLocation); ok && shopLocation.RadiusMeters > 0 {
		outShop = utils.IsWithinRadius(lat, lng, shopLocation.Lat, shopLocation.Lng, shopLocation.RadiusMeters)
	}

	var lbInShop, lbOutShop bool
	if lat, lng, ok := utils.ParseLocation(att.LunchBreakInLocation); ok && shopLocation.RadiusMeters > 0 {
		lbInShop = utils.IsWithinRadius(lat, lng, shopLocation.Lat, shopLocation.Lng, shopLocation.RadiusMeters)
	}
	if lat, lng, ok := utils.ParseLocation(att.LunchBreakOutLocation); ok && shopLocation.RadiusMeters > 0 {
		lbOutShop = utils.IsWithinRadius(lat, lng, shopLocation.Lat, shopLocation.Lng, shopLocation.RadiusMeters)
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
		StaffID:          encoder.Encode(att.StaffID),
		FullName:         utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
		Date:             att.ForDate,
		ScheduledTimeIn:  schedIn,
		ScheduledTimeOut: schedOut,

		Attendance: models.AttendanceStat{
			In:            timeIn,
			Out:           timeOut,
			InStatus:      c.timeInStatus,
			OutStatus:     c.timeOutStatus,
			Duration:      c.duration,
			DurationColor: c.durationColor,
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
			InStatus:      lunchbreak.timeInStatus,
			OutStatus:     lunchbreak.timeOutStatus,
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

func AddAdminHandlers(s *Server, r chi.Router) {
	r.Get("/admin", s.adminLoginPageHandler)
	r.Post("/admin/login", s.adminLoginHandler)
	r.With(s.requireStaffAuth).Post("/admin/logout", s.adminLogoutHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff", s.adminStaffHomeHandler)
	r.With(s.requireStaffAuth).Get("/admin/profile", s.adminStaffProfileHandler)
	r.With(s.requireStaffAuth).Get("/admin/profile-header", s.adminStaffProfileHeaderHandler)
	r.With(s.requireStaffAuth).Post("/admin/change-password", s.adminChangePasswordHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff/attendance", s.adminStaffPageHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff/attendance/table", s.adminStaffAttendanceTableHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff/attendance/rows", s.adminStaffAttendanceRowsHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/time-in", s.adminStaffTimeInHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/time-out", s.adminStaffTimeOutHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/lunch-break-start", s.adminStaffLunchBreakInHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/lunch-break-end", s.adminStaffLunchBreakOutHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff/time-off", s.adminStaffTimeOffPageHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/time-off", s.adminStaffTimeOffHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff/time-off/table", s.adminStaffTimeOffTableHandler)
	r.With(s.requireStaffAuth).Post("/admin/staff/attendance/location", s.adminStaffAttendanceLocationHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser", s.adminSuperuserHomeHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/attendance", s.adminSuperuserAttendancePageHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/attendance/table", s.adminSuperuserAttendanceHandler)
	r.With(s.requireSuperuserAuth).Post("/admin/superuser/attendance/report", s.adminSuperuserAttendanceReportHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/time-off", s.adminSuperuserTimeOffPageHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/time-off/table", s.adminSuperuserTimeOffTableHandler)
	r.With(s.requireSuperuserAuth).Patch("/admin/superuser/time-off/{id}/approve", s.adminSuperuserTimeOffApproveHandler)
	r.With(s.requireSuperuserAuth).Patch("/admin/superuser/time-off/{id}/cancel", s.adminSuperuserTimeOffCancelHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/products", s.adminSuperuserProductsHandler)
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
