package server

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/services"
	staffmodels "cchoice/internal/staff"
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
		redirectHXLogin(w, r)
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

	startDateParam := r.URL.Query().Get("date-selector")
	startDate := utils.ParseAttendanceDate(startDateParam)

	endDateParam := r.URL.Query().Get("date-selector-end")
	if endDateParam == "" {
		endDateParam = startDateParam
	}
	endDate := utils.ParseAttendanceDate(endDateParam)

	staffID := encode.INVALID
	if staffIDParam := r.FormValue("staff-id"); staffIDParam != "" {
		staffID = s.encoder.Decode(staffIDParam)
		if staffID == encode.INVALID {
			logs.Log().Warn(logtag, zap.String("staff id", staffIDParam), zap.Error(errs.ErrDecode))
		}
	}

	//TODO: (Brandon) services should be in the Server struct as well
	//                see how I implemented ProductsService in server
	attendanceService := services.NewAttendanceService(s.encoder, s.dbRO, s.dbRW)
	attendances, err := attendanceService.GetAttendance(ctx, staffID, startDate, endDate)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("start date", startDateParam), zap.String("end date", endDateParam))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

	attendanceData := make([]models.Attendance, 0, len(attendances))
	for _, att := range attendances {
		staff, ok := staffMap[att.StaffID]
		if !ok {
			continue
		}
		attendanceData = append(
			attendanceData,
			attendanceService.ComputeData(staffmodels.StaffRowBase(staff), att),
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

	if err := compadmin.AdminSuperuserAttendancePage("Employee Attendance", date).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserAttendanceReportHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Attendance Report Handler]"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/attendance", err.Error()))
		return
	}

	startDate := r.FormValue("date-selector")
	endDate := r.FormValue("date-selector-end")
	if startDate == "" || endDate == "" {
		redirectHX(w, r, utils.URLWithError("/admin/superuser/attendance", "missing start date or end date"))
		return
	}

	staffID := encode.INVALID
	if staffIDStr := r.FormValue("staff-id"); staffIDStr != "" {
		staffID = s.encoder.Decode(staffIDStr)
		if staffID == encode.INVALID {
			logs.Log().Warn(logtag, zap.String("staff id", staffIDStr), zap.Error(errs.ErrDecode))
		}
	}

	format := r.URL.Query().Get("format")

	//TODO: (Brandon) Create enums OUTPUT_FORMAT for csv, xlsx, etc values
	if format == "" {
		format = "csv"
	}

	attendanceService := services.NewAttendanceService(s.encoder, s.dbRO, s.dbRW)
	attendances, err := attendanceService.GetAttendance(ctx, staffID, startDate, endDate)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError("/admin/superuser/attendance", err.Error()))
		return
	}
	if len(attendances) == 0 {
		redirectHX(w, r, utils.URLWithError("/admin/superuser/attendance", "No attendance data found. Skipping report generation."))
		return
	}

	reportName := fmt.Sprintf("attendance_%s_%s_%s.%s", startDate, endDate, utils.GenString(8), format)
	w.Header().Set("Content-Disposition", "attachment; filename="+reportName)

	logs.Log().Info(
		logtag,
		zap.String("file", reportName),
		zap.String("start date", startDate),
		zap.String("end date", endDate),
		zap.Int64("staff id", s.sessionManager.GetInt64(ctx, SessionStaffID)),
	)

	reportService := services.NewReportService(s.encoder, s.dbRO)
	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		writer := csv.NewWriter(w)
		defer writer.Flush()

		if err := reportService.StreamReportCSV(
			ctx,
			writer,
			attendances,
			staffID,
			reportName,
			startDate,
			endDate,
		); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError("/admin/superuser/attendance", err.Error()))
			return
		}

		if err := writer.Error(); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError("/admin/superuser/attendance", err.Error()))
			return
		}
	case "xlsx":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		file := excelize.NewFile()
		defer file.Close()

		if err := reportService.StreamReportXLSX(
			ctx,
			file,
			attendances,
			staffID,
			reportName,
			startDate,
			endDate,
		); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError("/admin/superuser/attendance", err.Error()))
			return
		}

		if err := file.Write(w); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError("/admin/superuser/attendance", err.Error()))
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
		redirectHXLogin(w, r)
		return
	}

	timeOffIDStr := chi.URLParam(r, "id")
	decodedTimeOffID := s.encoder.Decode(timeOffIDStr)
	if decodedTimeOffID == encode.INVALID {
		logs.LogCtx(ctx).Error(logtag, zap.String("time_off_id", timeOffIDStr))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	attendanceService := services.NewAttendanceService(s.encoder, s.dbRO, s.dbRW)
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
	redirectHX(w, r, utils.URLWithSuccess("/admin/superuser/time-off", "Time off request approved"))
}

func (s *Server) adminSuperuserTimeOffCancelHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Time Off Cancel Handler]"
	ctx := r.Context()
	currentStaffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	_, err := s.dbRO.GetQueries().GetStaffByID(ctx, currentStaffID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Int64("staff_id", currentStaffID), zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	timeOffIDStr := chi.URLParam(r, "id")
	decodedTimeOffID := s.encoder.Decode(timeOffIDStr)
	if decodedTimeOffID == encode.INVALID {
		logs.LogCtx(ctx).Error(logtag, zap.String("time_off_id", timeOffIDStr))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	attendanceService := services.NewAttendanceService(s.encoder, s.dbRO, s.dbRW)
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
	redirectHX(w, r, utils.URLWithSuccess("/admin/superuser/time-off", "Time off request cancelled"))
}

func (s *Server) adminSuperuserProductsListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products List Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminSuperuserProductsListPage("Products").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserProductsListTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products List Table Handler]"
	ctx := r.Context()

	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")

	products, err := s.dbRO.GetQueries().AdminGetProductsForListing(ctx, queries.AdminGetProductsForListingParams{
		Search: sql.NullString{String: search, Valid: search != ""},
		Status: sql.NullString{String: status, Valid: status != ""},
	})
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		products = []queries.AdminGetProductsForListingRow{}
	}

	productList := make([]models.AdminProductListItem, 0, len(products))
	for _, p := range products {
		productList = append(productList, models.AdminProductListItem{
			ID:          s.encoder.Encode(p.ID),
			Name:        p.Name,
			Serial:      p.Serial,
			Description: p.Description.String,
			Brand:       p.BrandName,
			Status:      enums.ParseProductStatusToEnum(p.Status),
			ImagePath:   p.ImagePath,
			CreatedAt:   p.CreatedAt.Format(constants.DateTimeLayoutISO),
			UpdatedAt:   p.UpdatedAt.Format(constants.DateTimeLayoutISO),
		})
	}

	if err := compadmin.AdminSuperuserProductsListTable(productList).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserProductsUpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Products Update Status Handler]"
	const page = "/admin/superuser/products"
	ctx := r.Context()

	productIDStr := chi.URLParam(r, "id")
	if productIDStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "Invalid product ID"))
		return
	}

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, "Invalid form submission"))
		return
	}

	statusStr := r.FormValue("status")
	if statusStr == "" {
		redirectHX(w, r, utils.URLWithError(page, "Status is required"))
		return
	}

	status := enums.ParseProductStatusToEnum(statusStr)
	if status == enums.PRODUCT_STATUS_UNDEFINED {
		redirectHX(w, r, utils.URLWithError(page, "Invalid status"))
		return
	}

	if err := s.services.product.UpdateProductStatus(ctx, productIDStr, status); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("product_id", productIDStr),
			zap.String("status", statusStr),
			zap.Error(err),
		)
		redirectHX(w, r, utils.URLWithError(page, "Failed to update product status"))
		return
	}

	staffID := s.sessionManager.GetInt64(ctx, SessionStaffID)
	if err := s.services.staffLog.CreateLog(
		ctx,
		staffID,
		"update status",
		"products",
		"success",
		nil,
	); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}

	redirectHX(w, r, utils.URLWithSuccess(page, "Product status updated successfully"))
}
