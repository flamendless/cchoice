package server

import (
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
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
