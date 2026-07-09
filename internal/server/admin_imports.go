package server

import (
	"net/http"
	"strconv"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminImportsPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Imports Page Handler]"
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

	if err := compadmin.AdminImportsPage(isSuperuser, roles, backURL).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(backURL, err.Error()))
	}
}

func (s *Server) adminImportsProductsModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Imports Products Modal Handler]"
	const page = "/admin/imports"
	ctx := r.Context()

	if err := compadmin.ProductsImportModal().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminImportsProductsPreviewHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Imports Products Preview Handler]"
	const page = "/admin/imports"
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(errs.ErrInvalidParams)))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, errs.ErrMissingField.Error()))
		return
	}
	defer file.Close()

	preview, sessionData, err := s.services.productBulkImport.PreviewFromReader(ctx, header.Filename, file)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	s.sessionManager.Put(ctx, skProductImportPreview, sessionData)

	if err := compadmin.ProductsImportPreview(*preview).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}

func (s *Server) adminImportsProductsApplyHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Imports Products Apply Handler]"
	const page = "/admin/imports"
	ctx := r.Context()

	var f forms.AdminImportsProductsApplyForm
	if err := httputil.BindPostForm(r, &f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, httputil.ErrorMessage(err)))
		return
	}

	sessionData, ok := s.sessionManager.Get(ctx, skProductImportPreview).(*services.ProductImportSessionData)
	if !ok || sessionData == nil {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrImportPreviewExpired.Error()))
		return
	}

	selectedLines := make([]int, 0, len(f.Lines))
	for _, lineStr := range f.Lines {
		line, err := strconv.Atoi(lineStr)
		if err != nil {
			continue
		}
		selectedLines = append(selectedLines, line)
	}

	if len(selectedLines) == 0 {
		redirectHX(w, r, utils.URLWithError(page, errs.ErrImportNoRowsSelected.Error()))
		return
	}

	adminStaffID := s.sessionManager.GetString(ctx, SessionStaffID)
	result, err := s.services.productBulkImport.ApplySelected(ctx, adminStaffID, sessionData, selectedLines)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
		return
	}

	s.sessionManager.Remove(ctx, skProductImportPreview)

	if err := compadmin.ProductsImportResult(*result).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(page, err.Error()))
	}
}
