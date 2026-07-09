package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/services"
	"cchoice/internal/utils"

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

	var q forms.AdminSuperuserStaffsListQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}

	staffRows, err := s.services.staff.GetAllForAdmin(ctx, q.Search)
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
			Status:   enums.ParseStaffStatusToEnum(staff.Status),
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

	var q forms.AdminSuperuserStaffRolesQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		http.Error(w, "staff_id is required", http.StatusBadRequest)
		return
	}
	staffID := q.StaffID

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

	var p forms.AdminSuperuserStaffRolePath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrStaffIDRequired.Error()))
		return
	}
	staffID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrStaffIDRequired.Error()))
		return
	}

	var q forms.AdminSuperuserStaffRoleActionQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}
	action := q.Action

	var f forms.AdminSuperuserStaffRoleForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	role := enums.ParseStaffRoleToEnum(f.Role)
	if !role.IsValid() {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrEnumInvalid.Error()))
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
		redirectHX(w, r, utils.URLWithError(page, errs.ErrEnumInvalid.Error()))
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
	const listPage = "/admin/superuser/staffs"
	ctx := r.Context()

	var f forms.AdminSuperuserStaffCreateForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	if f.Password != f.ConfirmPassword {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrPasswordsDoNotMatch.Error()))
		return
	}

	userType := enums.ParseStaffUserTypeToEnum(f.UserType)
	if userType == enums.STAFF_USER_TYPE_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrEnumInvalid.Error()))
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

	createdStaffID, err := s.services.staff.Create(ctx, services.CreateStaffParams{
		FirstName:       f.FirstName,
		MiddleName:      f.MiddleName,
		LastName:        f.LastName,
		Birthdate:       f.Birthdate,
		Sex:             f.Sex,
		DateHired:       f.DateHired,
		Position:        f.Position,
		UserType:        userType,
		Email:           f.Email,
		MobileNo:        f.MobileNo,
		TimeInSchedule:  f.TimeInSchedule,
		TimeOutSchedule: f.TimeOutSchedule,
		Password:        f.Password,
		RequireInShop:   f.RequireInShop == "true",
		Status:          enums.ParseStaffStatusToEnum(f.Status),
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

	result = fmt.Sprintf("success. ID '%s'", createdStaffID)
	redirectHX(w, r, utils.URLWithSuccess(listPage, "Employee created successfully"))
}

func (s *Server) adminSuperuserStaffsEditPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Staffs Edit Page Handler]"
	const page = "/admin/superuser/staffs"
	ctx := r.Context()

	var p forms.AdminSuperuserStaffUpdatePath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrStaffIDRequired.Error()))
		return
	}
	staffID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrStaffIDRequired.Error()))
		return
	}

	staff, err := s.services.staff.GetForEdit(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		if errors.Is(err, sql.ErrNoRows) {
			redirectHX(w, r, utils.URLWithError(page, errs.ErrStaffNotFound.Error()))
			return
		}
		redirectHX(w, r, utils.URLWithError(page, errs.ErrStaffLoadFailed.Error()))
		return
	}

	if err := compadmin.StaffEditModal(staff).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrRenderFailed.Error()))
		return
	}
}

func (s *Server) adminSuperuserStaffsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Staffs Update Handler]"
	const page = "/admin/superuser/staffs"
	ctx := r.Context()

	var p forms.AdminSuperuserStaffUpdatePath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrStaffIDRequired.Error()))
		return
	}
	staffID, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrStaffIDRequired.Error()))
		return
	}

	var f forms.AdminSuperuserStaffUpdateForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	status := enums.ParseStaffStatusToEnum(f.Status)
	if !status.IsValid() {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrEnumInvalid.Error()))
		return
	}

	result := "success"
	defer func() {
		if err := s.services.staffLog.CreateLog(
			context.Background(),
			s.sessionManager.GetString(ctx, SessionStaffID),
			constants.ActionUpdate,
			constants.ModuleStaff,
			result,
			nil,
		); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
		}
	}()

	if err := s.services.staff.UpdateEmployment(ctx, services.UpdateEmploymentParams{
		ID:              staffID,
		Status:          status,
		Position:        f.Position,
		TimeInSchedule:  f.TimeInSchedule,
		TimeOutSchedule: f.TimeOutSchedule,
		RequireInShop:   f.RequireInShop == "true",
	}); err != nil {
		result = err.Error()
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		msg := "Failed to update employee"
		switch {
		case errors.Is(err, errs.ErrMissingField):
			msg = "All required fields must be filled"
		case errors.Is(err, errs.ErrInvalidFormat):
			msg = "Invalid format for one or more fields"
		case errors.Is(err, errs.ErrInvalidInput):
			msg = "Invalid input for one or more fields"
		case errors.Is(err, errs.ErrDecode):
			msg = "Invalid employee id"
		}
		redirectHX(w, r, utils.URLWithError(page, msg))
		return
	}

	result = fmt.Sprintf("success. ID '%s'", staffID)
	redirectHX(w, r, utils.URLWithSuccess(page, "Employee updated successfully"))
}
