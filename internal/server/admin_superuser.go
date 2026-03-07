package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminSuperuserHomeHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Home Handler]"
	ctx := r.Context()

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	staff, err := s.dbRO.GetQueries().GetStaffByID(ctx, staffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("staff_id", staffID), zap.Error(err))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}
	currentUserFullName := utils.BuildFullName(staff.FirstName, staff.MiddleName.String, staff.LastName)

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
		attendanceData = append(
			attendanceData,
			buildAdminStaffAttendance(s.encoder, staff, att, shop),
		)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserTimeOffPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Time Off Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminSuperuserTimeOffPage("Staff Time Off").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserTimeOffTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Time Off Table Handler]"
	ctx := r.Context()

	timeOffs, err := s.dbRO.GetQueries().GetAllStaffTimeOffs(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		timeOffs = []queries.GetAllStaffTimeOffsRow{}
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

		fullName := utils.BuildFullName(
			to.StaffFirstName,
			to.StaffMiddleName.String,
			to.StaffLastName,
		)

		staffTimeOffs = append(staffTimeOffs, models.StaffTimeOff{
			ID:          s.encoder.Encode(to.ID),
			StaffID:     s.encoder.Encode(to.StaffID),
			FullName:    fullName,
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

	if err := compadmin.AdminSuperuserTimeOffTable(staffTimeOffs).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserTimeOffApproveHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Time Off Approve Handler]"
	ctx := r.Context()
	currentStaffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	_, err := s.dbRO.GetQueries().GetStaffByID(ctx, currentStaffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("staff_id", currentStaffID), zap.Error(err))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}

	timeOffIDStr := chi.URLParam(r, "id")
	decodedTimeOffID := s.encoder.Decode(timeOffIDStr)
	if decodedTimeOffID == -1 {
		logs.LogCtx(ctx).Error(logtag, zap.String("time_off_id", timeOffIDStr))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	attendanceService := services.NewAttendanceService(s.dbRO, s.dbRW)
	if err := attendanceService.ApproveTimeOff(ctx, decodedTimeOffID, currentStaffID); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("time_off_id", timeOffIDStr),
			zap.Int64("time_off_id", decodedTimeOffID),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) adminSuperuserTimeOffCancelHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Time Off Cancel Handler]"
	ctx := r.Context()
	currentStaffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	_, err := s.dbRO.GetQueries().GetStaffByID(ctx, currentStaffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("staff_id", currentStaffID), zap.Error(err))
		http.Redirect(w, r, utils.URL("/admin"), http.StatusSeeOther)
		return
	}

	timeOffIDStr := chi.URLParam(r, "id")
	decodedTimeOffID := s.encoder.Decode(timeOffIDStr)
	if decodedTimeOffID == -1 {
		logs.LogCtx(ctx).Error(logtag, zap.String("time_off_id", timeOffIDStr))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	attendanceService := services.NewAttendanceService(s.dbRO, s.dbRW)
	if err := attendanceService.CancelTimeOff(ctx, decodedTimeOffID, currentStaffID); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("time_off_id", timeOffIDStr),
			zap.Int64("time_off_id", decodedTimeOffID),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}
