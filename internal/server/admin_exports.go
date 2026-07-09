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
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"

	"github.com/xuri/excelize/v2"
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

	var q forms.AdminExportsProductsCountQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}
	brand := q.Brand
	status := enums.ParseProductStatusToEnum(q.Status)

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

	var f forms.AdminExportsProductsForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeExportError(w, http.StatusBadRequest, err.Error())
		return
	}

	brand := f.Brand
	status := enums.ParseProductStatusToEnum(f.Status)
	sortColumn, sortDirection := parseProductExportSortParams(
		f.SortColumn,
		f.SortDirection,
	)

	formatEnum := enums.ParseOutputFormatToEnum(f.Format)
	if formatEnum == enums.OUTPUT_FORMAT_UNDEFINED {
		formatEnum = enums.OUTPUT_FORMAT_CSV
	}
	if formatEnum != enums.OUTPUT_FORMAT_CSV && formatEnum != enums.OUTPUT_FORMAT_XLSX {
		writeExportError(w, http.StatusBadRequest, "unsupported export format")
		return
	}

	adminStaffID := s.sessionManager.GetString(ctx, SessionStaffID)
	reportName := fmt.Sprintf(
		"export_products_%s.%s",
		time.Now().Format(constants.DateTimeLayoutFilename),
		strings.ToLower(formatEnum.String()),
	)

	switch formatEnum {
	case enums.OUTPUT_FORMAT_CSV:
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		if err := s.services.export.StreamProductsCSV(ctx, writer, brand, status, sortColumn, sortDirection, adminStaffID, reportName); err != nil {
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
	case enums.OUTPUT_FORMAT_XLSX:
		file := excelize.NewFile()
		defer file.Close()

		if err := s.services.export.StreamProductsXLSX(ctx, file, brand, status, sortColumn, sortDirection, adminStaffID, reportName); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			writeExportError(w, http.StatusInternalServerError, err.Error())
			return
		}

		var buf bytes.Buffer
		if err := file.Write(&buf); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			writeExportError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+reportName)
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		if _, err := w.Write(buf.Bytes()); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		}
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
