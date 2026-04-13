package server

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminSuperuserHomeHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Home Handler]"
	ctx := r.Context()

	staffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	staffProfile, err := s.services.staff.GetCurrentStaff(ctx, staffIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", staffProfile.ID), zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	if err := compadmin.AdminSuperuserHomePage(staffProfile.FullName).Render(ctx, w); err != nil {
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

	startDateParam := r.URL.Query().Get("date-selector")
	startDate := utils.ParseAttendanceDate(startDateParam)

	endDateParam := r.URL.Query().Get("date-selector-end")
	if endDateParam == "" {
		endDateParam = startDateParam
	}
	endDate := utils.ParseAttendanceDate(endDateParam)

	attendances, err := s.services.attendance.GetAttendance(ctx, r.FormValue("staff-id"), startDate, endDate)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("start date", startDateParam), zap.String("end date", endDateParam))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	attendanceData := s.services.attendance.ComputeAllAttendanceData(ctx, maxStaffListSize, attendances)

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

	if err := compadmin.AdminSuperuserAttendancePage("Employee Attendance", date).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserAttendanceReportHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Attendance Report Handler]"
	const page = "/admin/superuser/attendance"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	startDate := r.FormValue("date-selector")
	endDate := r.FormValue("date-selector-end")
	if startDate == "" || endDate == "" {
		redirectHX(w, r, utils.URLWithError(page, "missing start date or end date"))
		return
	}

	formatParam := r.URL.Query().Get("format")
	formatEnum := enums.ParseOutputFormatToEnum(formatParam)
	if formatEnum == enums.OUTPUT_FORMAT_UNDEFINED {
		formatEnum = enums.OUTPUT_FORMAT_CSV
	}

	staffID := r.FormValue("staff-id")
	adminStaffID := s.sessionManager.GetString(ctx, SessionStaffID)
	attendances, err := s.services.attendance.GetAttendance(ctx, staffID, startDate, endDate)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}
	if len(attendances) == 0 {
		redirectHX(w, r, utils.URLWithError(page, "No attendance data found. Skipping report generation."))
		return
	}

	reportName := fmt.Sprintf("attendance_%s_%s_%s.%s", startDate, endDate, utils.GenString(8), strings.ToLower(formatEnum.String()))
	w.Header().Set("Content-Disposition", "attachment; filename="+reportName)

	logs.Log().Info(
		logtag,
		zap.String("file", reportName),
		zap.String("start date", startDate),
		zap.String("end date", endDate),
		zap.String("staff id", adminStaffID),
		zap.String("param staff id", staffID),
	)

	switch formatEnum {
	case enums.OUTPUT_FORMAT_CSV:
		w.Header().Set("Content-Type", "text/csv")
		writer := csv.NewWriter(w)
		defer writer.Flush()

		if err := s.services.report.StreamReportCSV(
			ctx,
			writer,
			attendances,
			adminStaffID,
			staffID,
			reportName,
			startDate,
			endDate,
		); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}

		if err := writer.Error(); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}
	case enums.OUTPUT_FORMAT_XLSX:
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		file := excelize.NewFile()
		defer file.Close()

		if err := s.services.report.StreamReportXLSX(
			ctx,
			file,
			attendances,
			adminStaffID,
			staffID,
			reportName,
			startDate,
			endDate,
		); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}

		if err := file.Write(w); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(page, err.Error()))
			return
		}
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

	staffTimeOffs, err := s.services.attendance.GetAllStaffTimeOffs(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		staffTimeOffs = []models.StaffTimeOff{}
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
	currentStaffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	staff, err := s.services.staff.GetCurrentStaff(ctx, currentStaffIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", staff.ID), zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	timeOffID := chi.URLParam(r, "id")
	if err := s.services.attendance.ApproveTimeOff(ctx, timeOffID, currentStaffIDStr); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("staff_id", currentStaffIDStr),
			zap.String("time_off_id", timeOffID),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	redirectHX(w, r, utils.URLWithSuccess("/admin/superuser/time-off", "Time off request approved"))
}

func (s *Server) adminSuperuserTimeOffCancelHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Time Off Cancel Handler]"
	ctx := r.Context()
	currentStaffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	staff, err := s.services.staff.GetCurrentStaff(ctx, currentStaffIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("staff_id", staff.ID), zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	timeOffID := chi.URLParam(r, "id")
	if err := s.services.attendance.CancelTimeOff(ctx, timeOffID, currentStaffIDStr); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("staff_id", currentStaffIDStr),
			zap.String("time_off_id", timeOffID),
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	redirectHX(w, r, utils.URLWithSuccess("/admin/superuser/time-off", "Time off request cancelled"))
}

func (s *Server) adminCustomersListHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Customers List Handler]"
	ctx := r.Context()

	customers, err := s.dbRO.GetQueries().GetAllCustomers(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return
	}

	opts := make([]struct {
		ID    string
		Email string
	}, len(customers))
	for i, c := range customers {
		opts[i] = struct {
			ID    string
			Email string
		}{
			ID:    s.encoder.Encode(c.ID),
			Email: c.Email,
		}
	}

	if err := compadmin.CustomerOptions(opts).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}
