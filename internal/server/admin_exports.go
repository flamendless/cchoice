package server

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminExportsPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Exports Page Handler]"
	ctx := r.Context()

	staffIDStr := s.sessionManager.GetString(ctx, SessionStaffID)
	staff, err := s.services.staff.GetCurrentStaff(ctx, staffIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHXLogin(w, r)
		return
	}

	roles, err := s.services.role.GetByStaffID(ctx, staffIDStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}

	isSuperuser := staff.UserType == enums.STAFF_USER_TYPE_SUPERUSER.String()
	backURL := "/admin/staff"
	if isSuperuser {
		backURL = "/admin/superuser"
	}

	if err := compadmin.AdminExportsPage(isSuperuser, roles, backURL).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(backURL, err.Error()))
	}
}

func (s *Server) adminExportsProductsModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Exports Products Modal Handler]"
	const page = "/admin/exports"
	ctx := r.Context()

	brandsRes, err := requests.GetBrandsForAdmin(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		requests.GenerateAdminBrandsCacheKey(),
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		brandsRes = nil
	}

	brands := make([]models.AdminBrand, 0, len(brandsRes))
	for _, b := range brandsRes {
		brands = append(brands, models.AdminBrand{
			ID:   s.encoder.Encode(b.ID),
			Name: b.Name,
		})
	}

	if err := compadmin.ProductsExportModal(brands).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminExportsProductsCountHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Exports Products Count Handler]"
	ctx := r.Context()

	brand := r.URL.Query().Get("brand")
	status := enums.ParseProductStatusToEnum(r.URL.Query().Get("status"))

	count, err := s.services.product.CountForExportAdmin(ctx, brand, status)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := compadmin.ProductsExportCount(count).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}

func (s *Server) adminExportsProductsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Exports Products Handler]"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeExportError(w, http.StatusBadRequest, err.Error())
		return
	}

	brand := r.PostFormValue("brand")
	status := enums.ParseProductStatusToEnum(r.PostFormValue("status"))
	sortColumn, sortDirection := parseProductExportSortParams(
		r.PostFormValue("sort_column"),
		r.PostFormValue("sort_direction"),
	)

	formatEnum := enums.ParseOutputFormatToEnum(r.PostFormValue("format"))
	if formatEnum == enums.OUTPUT_FORMAT_UNDEFINED {
		formatEnum = enums.OUTPUT_FORMAT_CSV
	}
	if formatEnum != enums.OUTPUT_FORMAT_CSV {
		writeExportError(w, http.StatusBadRequest, "unsupported export format")
		return
	}

	adminStaffID := s.sessionManager.GetString(ctx, SessionStaffID)
	reportName := fmt.Sprintf(
		"export_products_%s.%s",
		time.Now().Format(constants.DateTimeLayoutFilename),
		strings.ToLower(formatEnum.String()),
	)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := s.services.export.StreamProductsCSV(ctx, writer, brand, status, sortColumn, sortDirection, adminStaffID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeExportError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeExportError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+reportName)
	w.Header().Set("Content-Type", "text/csv")
	if _, err := w.Write(buf.Bytes()); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}

func writeExportError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func parseProductExportSortParams(sortColumnRaw, sortDirectionRaw string) (enums.ProductExportSortColumn, enums.ProductExportSortDirection) {
	sortColumn := enums.ParseProductExportSortColumnToEnum(sortColumnRaw)
	if sortColumn == enums.PRODUCT_EXPORT_SORT_COLUMN_UNDEFINED {
		sortColumn = enums.PRODUCT_EXPORT_SORT_COLUMN_UPDATED_AT
	}

	sortDirection := enums.ParseProductExportSortDirectionToEnum(sortDirectionRaw)
	if sortDirection == enums.PRODUCT_EXPORT_SORT_DIRECTION_UNDEFINED {
		sortDirection = enums.PRODUCT_EXPORT_SORT_DIRECTION_DESC
	}

	return sortColumn, sortDirection
}
