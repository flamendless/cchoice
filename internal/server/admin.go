package server

import (
	"context"
	"database/sql"
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
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

func getOrCreateUserAgentID(ctx context.Context, db database.IService, userAgentStr string) sql.NullInt64 {
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

func AddAdminHandlers(s *Server, r chi.Router) {
	r.Get("/admin", s.adminLoginPageHandler)
	r.Post("/admin/login", s.adminLoginHandler)
	r.With(s.requireStaffAuth).Post("/admin/logout", s.adminLogoutHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff", s.adminStaffHomeHandler)
	r.With(s.requireStaffAuth).Get("/admin/profile", s.adminStaffProfileHandler)
	r.With(s.requireStaffAuth).Get("/admin/profile-header", s.adminStaffProfileHeaderHandler)
	r.With(s.requireStaffAuth).Get("/admin/profile/edit", s.adminProfileEditFormHandler)
	r.With(s.requireStaffAuth).Patch("/admin/profile", s.adminProfileUpdateHandler)
	r.With(s.requireStaffAuth).Post("/admin/change-password", s.adminChangePasswordHandler)
	r.With(s.requireStaffAuth).Get("/admin/staff/list", s.adminStaffListHandler)
	r.With(s.requireStaffAuth).Get("/admin/customers/list", s.adminCustomersListHandler)
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

	r.With(s.requireStaffAuth, s.AllowRoles(enums.STAFF_ROLE_CREATE_PRODUCT)).Get("/admin/superuser/products/create", s.adminSuperuserProductsCreatePageHandler)
	r.With(s.requireStaffAuth, s.AllowRoles(enums.STAFF_ROLE_CREATE_PRODUCT)).Post("/admin/superuser/products/create", s.adminSuperuserProductsCreatePostHandler)
	r.With(s.requireStaffAuth, s.AllowRoles(enums.STAFF_ROLE_CREATE_PRODUCT)).Get("/admin/superuser/products/subcategories", s.adminSuperuserProductsSubcategoriesHandler)
	r.With(s.requireStaffAuth, s.AllowRoles(enums.STAFF_ROLE_CREATE_PRODUCT)).Get("/admin/superuser/products/validate-serial", s.adminSuperuserProductsValidateSerialHandler)

	r.With(s.requireStaffAuth, s.AllowRoles(enums.STAFF_ROLE_CREATE_CPOINTS)).Get("/admin/cpoints/generate", s.adminCPointsGeneratePageHandler)
	r.With(s.requireStaffAuth, s.AllowRoles(enums.STAFF_ROLE_CREATE_CPOINTS)).Post("/admin/cpoints/generate", s.adminCPointsGeneratePostHandler)
	r.With(s.requireStaffAuth, s.AllowRoles(enums.STAFF_ROLE_CREATE_CPOINTS)).Get("/admin/cpoints/code", s.adminCPointsCodePageHandler)
	r.With(s.requireStaffAuth, s.AllowRoles(enums.STAFF_ROLE_CREATE_CPOINTS)).Get("/admin/cpoints/qr", s.adminCPointsQRHandler)

	r.With(s.requireSuperuserAuth).Get("/admin/superuser", s.adminSuperuserHomeHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/attendance", s.adminSuperuserAttendancePageHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/attendance/table", s.adminSuperuserAttendanceHandler)
	r.With(s.requireSuperuserAuth).Post("/admin/superuser/attendance/report", s.adminSuperuserAttendanceReportHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/time-off", s.adminSuperuserTimeOffPageHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/time-off/table", s.adminSuperuserTimeOffTableHandler)
	r.With(s.requireSuperuserAuth).Patch("/admin/superuser/time-off/{id}/approve", s.adminSuperuserTimeOffApproveHandler)
	r.With(s.requireSuperuserAuth).Patch("/admin/superuser/time-off/{id}/cancel", s.adminSuperuserTimeOffCancelHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/products", s.adminSuperuserProductsListPageHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/products/table", s.adminSuperuserProductsListTableHandler)
	r.With(s.requireSuperuserAuth).Patch("/admin/superuser/products/{id}/status", s.adminSuperuserProductsUpdateStatusHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/logs", s.adminSuperuserLogsPageHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/logs/table", s.adminSuperuserLogsTableHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/staffs", s.adminSuperuserStaffsListPageHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/staffs/table", s.adminSuperuserStaffsListTableHandler)
	r.With(s.requireSuperuserAuth).Get("/admin/superuser/staffs/roles", s.adminSuperuserStaffsRolesOptionsHandler)
	r.With(s.requireSuperuserAuth).Patch("/admin/superuser/staffs/{id}/role", s.adminSuperuserStaffsRoleHandler)
}

func (s *Server) adminLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Login Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminLoginPage().Render(ctx, w); err != nil {
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
		redirectHX(w, r, utils.URLWithError("/admin", "Invalid form submission"))
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	if !constants.ReEmail.MatchString(email) {
		redirectHX(w, r, utils.URLWithError("/admin", "Invalid email or password format"))
		return
	}

	if !constants.RePassword.MatchString(password) {
		redirectHX(w, r, utils.URLWithError("/admin", "Invalid email or password format"))
		return
	}

	staff, err := s.dbRO.GetQueries().GetStaffByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			redirectHX(w, r, utils.URLWithError("/admin", "Invalid email or password"))
			return
		}
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.Password), []byte(password)); err != nil {
		redirectHX(w, r, utils.URLWithError("/admin", "Invalid email or password"))
		return
	}

	s.sessionManager.Put(ctx, SessionStaffID, s.encoder.Encode(staff.ID))

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
		redirectHX(w, r, utils.URL("/admin/superuser"))
	case enums.STAFF_USER_TYPE_STAFF:
		redirectHX(w, r, utils.URL("/admin/staff"))
	default:
		logs.Log().Warn(logtag, zap.String("got unhandled", staff.UserType))
		redirectHX(w, r, utils.URL("/admin"))
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

	redirectHXLogin(w, r)
}
