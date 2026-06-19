package server

import (
	"context"
	"errors"
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *Server) adminSuperuserStaffsListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Staffs List Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminSuperuserStaffsListPage("Employees").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserStaffsListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Staffs List Table Handler]"
	ctx := r.Context()

	search := r.URL.Query().Get("search")

	staffRows, err := s.services.staff.GetAllForAdmin(ctx, search)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	staffList := make([]models.AdminStaffListItem, 0, len(staffRows))
	for _, staff := range staffRows {
		roles, err := s.services.role.GetByStaffID(ctx, s.encoder.Encode(staff.ID))
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			roles = []enums.StaffRole{}
		}

		staffList = append(staffList, models.AdminStaffListItem{
			ID:       s.encoder.Encode(staff.ID),
			FullName: utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName),
			Position: staff.Position,
			Email:    staff.Email,
			MobileNo: staff.MobileNo,
			UserType: enums.ParseStaffUserTypeToEnum(staff.UserType),
			Roles:    roles,
		})
	}

	if err := compadmin.AdminSuperuserStaffsListTable(staffList).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserStaffsRolesOptionsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Staffs Roles Options Handler]"
	ctx := r.Context()

	staffID := r.URL.Query().Get("staff_id")
	if staffID == "" {
		http.Error(w, "staff_id is required", http.StatusBadRequest)
		return
	}

	existingRoles, err := s.services.role.GetByStaffID(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	existingRoleSet := make(map[enums.StaffRole]bool)
	for _, role := range existingRoles {
		existingRoleSet[role] = true
	}

	var availableRoles []enums.StaffRole
	for _, role := range enums.GetAllStaffRoles() {
		if !existingRoleSet[role] {
			availableRoles = append(availableRoles, role)
		}
	}

	if err := compadmin.StaffRoleDropdown(staffID, availableRoles).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserStaffsRoleHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Staffs Role Handler]"
	const page = "/admin/superuser/staffs"
	ctx := r.Context()

	staffID := chi.URLParam(r, "id")
	if staffID == "" {
		redirectHX(w, r, utils.URLWithError(page, "staff id is required"))
		return
	}

	action := r.URL.Query().Get("action")
	if action == "" {
		redirectHX(w, r, utils.URLWithError(page, "action in required"))
		return
	}

	if err := r.ParseForm(); err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	role := enums.ParseStaffRoleToEnum(r.FormValue("role"))
	if !role.IsValid() {
		redirectHX(w, r, utils.URLWithError(page, "invalid role"))
		return
	}

	switch action {
	case "ADD":
		if err := s.services.role.AddRole(ctx, staffID, role); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Stringer("role", role), zap.String("action", action), zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}

		roles, err := s.services.role.GetByStaffID(ctx, staffID)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Stringer("role", role), zap.String("action", action), zap.Error(err))
			roles = []enums.StaffRole{}
		}

		userType := enums.STAFF_USER_TYPE_STAFF
		staffRows, err := s.services.staff.GetAllForAdmin(ctx, "")
		if err == nil {
			for _, staff := range staffRows {
				if s.encoder.Encode(staff.ID) == staffID {
					userType = enums.ParseStaffUserTypeToEnum(staff.UserType)
					break
				}
			}
		}

		if err := compadmin.StaffRolesCell(staffID, userType, roles).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Stringer("role", role), zap.String("action", action), zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}

	case "REMOVE":
		if err := s.services.role.RemoveRole(ctx, staffID, role); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Stringer("role", role), zap.String("action", action), zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}

	default:
		redirectHX(w, r, utils.URLWithError(page, "invalid action"))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "successfully updated roles of staff"))
}

func (s *Server) adminSuperuserStaffsCreatePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Staffs Create Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminSuperuserStaffsCreatePage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserStaffsCreatePostHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Staffs Create Post Handler]"
	const page = "/admin/superuser/staffs/create"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	if password != confirmPassword {
		redirectHX(w, r, utils.URLWithError(page, "Passwords do not match"))
		return
	}

	userType := enums.ParseStaffUserTypeToEnum(r.FormValue("user_type"))
	if userType == enums.STAFF_USER_TYPE_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(page, "Invalid user type"))
		return
	}

	result := "success"
	defer func() {
		if err := s.services.staffLog.CreateLog(
			context.Background(),
			s.sessionManager.GetString(ctx, SessionStaffID),
			constants.ActionCreate,
			constants.ModuleStaff,
			result,
			nil,
		); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
		}
	}()

	_, err := s.services.staff.Create(ctx, services.CreateStaffParams{
		FirstName:       r.FormValue("first_name"),
		MiddleName:      r.FormValue("middle_name"),
		LastName:        r.FormValue("last_name"),
		Birthdate:       r.FormValue("birthdate"),
		Sex:             r.FormValue("sex"),
		DateHired:       r.FormValue("date_hired"),
		Position:        r.FormValue("position"),
		UserType:        userType,
		Email:           r.FormValue("email"),
		MobileNo:        r.FormValue("mobile_no"),
		TimeInSchedule:  r.FormValue("time_in_schedule"),
		TimeOutSchedule: r.FormValue("time_out_schedule"),
		Password:        password,
		RequireInShop:   r.FormValue("require_in_shop") == "true",
	})
	if err != nil {
		result = err.Error()
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		msg := "Failed to create employee"
		switch {
		case errors.Is(err, errs.ErrDuplicateEmail):
			msg = "Email already exists"
		case errors.Is(err, errs.ErrMissingField):
			msg = "All required fields must be filled"
		case errors.Is(err, errs.ErrInvalidFormat):
			msg = "Invalid format for one or more fields"
		case errors.Is(err, errs.ErrInvalidInput):
			msg = "Invalid input for one or more fields"
		case errors.Is(err, errs.ErrValidationInvalidMobileNumber):
			msg = "Invalid mobile number format"
		}
		redirectHX(w, r, utils.URLWithError(page, msg))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess("/admin/superuser/staffs", "Employee created successfully"))
}
