package server

import (
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/database/queries"
	"cchoice/internal/logs"
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
