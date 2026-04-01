package server

import (
	"database/sql"
	"net/http"
	"time"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/services"
	"cchoice/internal/types"
	"cchoice/internal/utils"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

func (s *Server) adminStaffHomeHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Home Handler]"
	ctx := r.Context()

	staffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	staff, err := s.services.staff.GetCurrentStaff(ctx, staffIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", staff.ID), zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	roles, err := s.services.role.GetByStaffID(ctx, staffIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", staff.ID), zap.Error(err))
	}

	if err := compadmin.AdminStaffHomePage(staff.FullName, roles).Render(ctx, w); err != nil {
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

	profile, err := s.services.staff.GetCurrentStaff(ctx, s.sessionManager.GetString(ctx, SessionStaffID))
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", profile.ID), zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	if err := compadmin.AdminProfilePage(profile).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffProfileHeaderHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Profile Header Handler]"
	ctx := r.Context()

	profile, err := s.services.staff.GetCurrentStaff(ctx, s.sessionManager.GetString(ctx, SessionStaffID))
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", profile.ID), zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	if err := compadmin.AdminProfileHeader(profile).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Change Password Handler]"
	const page = "/admin/profile"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	newPassword := r.PostFormValue("new_password")
	confirmPassword := r.PostFormValue("confirm_password")

	if newPassword == "" || confirmPassword == "" {
		redirectHX(w, r, utils.URLWithError(page, "Both password fields are required"))
		return
	}

	if !constants.RePassword.MatchString(newPassword) {
		redirectHX(w, r, utils.URLWithError(page, "Invalid password format"))
		return
	}

	if newPassword != confirmPassword {
		redirectHX(w, r, utils.URLWithError(page, "Passwords do not match"))
		return
	}

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	if err := s.services.staff.UpdatePassword(ctx, staffID, newPassword); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("staff_id", staffID))
		redirectHX(w, r, utils.URLWithError(page, "Failed to update password"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Password updated successfully"))
}

func (s *Server) adminProfileEditFormHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Profile Edit Form Handler]"
	ctx := r.Context()

	profile, err := s.services.staff.GetCurrentStaff(ctx, s.sessionManager.GetString(ctx, SessionStaffID))
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", profile.ID), zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	if err := compadmin.AdminProfileEditForm(profile).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.String("staff_id", profile.ID),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Profile Update Handler]"
	const page = "/admin/profile"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	firstName := r.PostFormValue("first_name")
	middleName := r.PostFormValue("middle_name")
	lastName := r.PostFormValue("last_name")
	mobileNo := r.PostFormValue("mobile_no")
	birthdate := r.PostFormValue("birthdate")
	dateHired := r.PostFormValue("date_hired")

	if firstName == "" || lastName == "" || mobileNo == "" || birthdate == "" || dateHired == "" {
		redirectHX(w, r, utils.URLWithError(page, "All required fields must be filled"))
		return
	}

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	if err := s.services.staff.UpdateProfile(ctx, services.UpdateProfileParams{
		ID:         staffID,
		FirstName:  firstName,
		MiddleName: middleName,
		LastName:   lastName,
		MobileNo:   mobileNo,
		Birthdate:  birthdate,
		DateHired:  dateHired,
	}); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("staff_id", staffID))
		redirectHX(w, r, utils.URLWithError(page, "Failed to update profile"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Profile updated successfully"))
}

func (s *Server) adminStaffListHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff List Handler]"
	ctx := r.Context()

	staff, err := s.services.staff.GetCurrentStaff(ctx, s.sessionManager.GetString(ctx, SessionStaffID))
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", staff.ID), zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	var list []models.Staff
	switch staff.UserType {
	case enums.STAFF_USER_TYPE_STAFF:
		list = append(list, models.Staff{
			ID:       staff.ID,
			FullName: staff.FullName,
		})
	case enums.STAFF_USER_TYPE_SUPERUSER:
		list, err = s.services.staff.GetAll(ctx, 100)
		if err != nil {
			logs.Log().Error(logtag, zap.String("staff_id", staff.ID), zap.Error(err))
		}
	}
	if err := compadmin.StaffOptions(list).Render(ctx, w); err != nil {
		logs.Log().Error(logtag, zap.String("staff_id", staff.ID), zap.Error(err))
	}
}

func (s *Server) adminStaffPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Page Handler]"
	ctx := r.Context()

	staffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	profile, err := s.services.staff.GetCurrentStaffWithAttendance(ctx, staffIDStr, s.sessionManager)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", profile.ID), zap.Error(err))
		redirectHXLogin(w, r)
		return
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
	const logtag = "[Admin Staff Attendance Table Handler]"
	ctx := r.Context()

	staffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	if _, err := s.services.staff.GetCurrentStaff(ctx, staffIDStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", staffIDStr), zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	date := utils.ParseAttendanceDate(r.URL.Query().Get("date"))
	dayAtt, err := s.services.attendance.GetStaffDayAttendance(ctx, staffIDStr, date)
	var record *models.Attendance
	if err == nil {
		record = dayAtt.Computed
	}

	if err := compadmin.StaffAttendanceSingleTable(record).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffAttendanceRowsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Attendance Rows Handler]"
	ctx := r.Context()

	staffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	if _, err := s.services.staff.GetCurrentStaff(ctx, staffIDStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", staffIDStr), zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	date := utils.ParseAttendanceDate(r.URL.Query().Get("date-selector"))
	dayAtt, err := s.services.attendance.GetStaffDayAttendance(ctx, staffIDStr, date)
	var record *models.Attendance
	if err == nil {
		record = dayAtt.Computed
	}

	if err := compadmin.StaffAttendanceRows(record).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffTimeInHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now().UTC().Format(constants.DateTimeLayoutISO)
	date := utils.NowPH().Format(constants.DateLayoutISO)
	location := GetLocation(ctx, s.sessionManager)
	useragentID := getOrCreateUserAgentID(ctx, s.dbRW, r.UserAgent())
	if err := s.services.attendance.TimeIn(ctx, s.sessionManager.GetString(ctx, SessionStaffID), date, now, location, useragentID); err != nil {
		http.Error(w, "Unable to time in", http.StatusBadRequest)
		return
	}
	redirectHX(w, r, utils.URLWithSuccess("/admin/staff/attendance", "Time in recorded successfully"))
}

func (s *Server) adminStaffTimeOutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now().UTC().Format(constants.DateTimeLayoutISO)
	date := utils.NowPH().Format(constants.DateLayoutISO)
	location := GetLocation(ctx, s.sessionManager)
	useragentID := getOrCreateUserAgentID(ctx, s.dbRW, r.UserAgent())
	if err := s.services.attendance.TimeOut(ctx, s.sessionManager.GetString(ctx, SessionStaffID), date, now, location, useragentID); err != nil {
		http.Error(w, "Unable to time out", http.StatusBadRequest)
		return
	}
	redirectHX(w, r, utils.URLWithSuccess("/admin/staff/attendance", "Time out recorded successfully"))
}

func (s *Server) adminStaffLunchBreakInHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Lunch Break Start Handler]"
	ctx := r.Context()
	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	now := time.Now().UTC().Format(constants.DateTimeLayoutISO)
	date := utils.NowPH().Format(constants.DateLayoutISO)
	location := GetLocation(ctx, s.sessionManager)
	useragentID := getOrCreateUserAgentID(ctx, s.dbRW, r.UserAgent())
	if err := s.services.attendance.LunchBreakIn(ctx, staffID, date, now, location, useragentID); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
			zap.String("staff_id", staffID),
			zap.String("date", date),
		)
		http.Error(w, "Unable to lunchbreak in", http.StatusBadRequest)
		return
	}
	redirectHX(w, r, utils.URLWithSuccess("/admin/staff/attendance", "Lunch break start recorded successfully"))
}

func (s *Server) adminStaffLunchBreakOutHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Lunch Break End Handler]"
	ctx := r.Context()
	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	now := time.Now().UTC().Format(constants.DateTimeLayoutISO)
	date := utils.NowPH().Format(constants.DateLayoutISO)
	location := GetLocation(ctx, s.sessionManager)
	useragentID := getOrCreateUserAgentID(ctx, s.dbRW, r.UserAgent())
	if err := s.services.attendance.LunchBreakOut(ctx, staffID, date, now, location, useragentID); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
			zap.String("staff_id", staffID),
			zap.String("date", date),
		)
		http.Error(w, "Unable to lunchbreak out", http.StatusBadRequest)
		return
	}
	redirectHX(w, r, utils.URLWithSuccess("/admin/staff/attendance", "Lunch break end recorded successfully"))
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
	if err := s.services.attendance.TimeOff(
		ctx,
		s.sessionManager.GetString(ctx, SessionStaffID),
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
	redirectHX(w, r, utils.URLWithSuccess("/admin/staff/time-off", "Time off request submitted"))
}

func (s *Server) adminStaffTimeOffTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Staff Time Off Table Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	if staffID == "" {
		redirectHXLogin(w, r)
		return
	}

	staffTimeOffs, err := s.services.staff.GetTimeOffs(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", staffID), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := compadmin.StaffTimeOffsTable(staffTimeOffs).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminStaffAttendanceLocationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	staffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	profile, err := s.services.staff.GetCurrentStaff(ctx, staffIDStr)
	if err != nil {
		redirectHXLogin(w, r)
		return
	}

	lat, lng, err := s.services.location.ComputeLocationFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	locationJSON, _ := json.Marshal(types.Location{Lat: lat, Lng: lng})
	attendanceLocation := sql.NullString{String: string(locationJSON), Valid: true}

	locationResult := s.services.location.ComputeLocation(attendanceLocation, GetLocation(ctx, s.sessionManager))

	profile.InShop = locationResult.InShop
	profile.LocationDisplay = locationResult.LocationDisplay
	profile.DistanceMeters = locationResult.DistanceMeters

	_ = compadmin.StaffLocationStatus(profile).Render(ctx, w)
}
