package server

import (
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminSuperuserLogsPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Logs Page Handler]"
	ctx := r.Context()

	if err := compadmin.AdminSuperuserLogsPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminSuperuserLogsTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Logs Table Handler]"
	ctx := r.Context()

	var q forms.AdminLogsFilterQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}
	page := httputil.PageOrDefault(q.Page, 1)

	var staffID int64
	if q.StaffID != "" {
		decoded := s.encoder.Decode(q.StaffID)
		if decoded != encode.INVALID {
			staffID = decoded
		}
	}

	module := enums.ParseModuleToEnum(q.Module)

	logsList, totalCount, page, err := s.services.staffLog.GetFilteredAsModelPaginated(
		ctx,
		staffID,
		q.Action,
		module,
		page,
		constants.DefaultAdminTablePageSize,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pagination := models.TablePagination{
		Page:          page,
		PerPage:       constants.DefaultAdminTablePageSize,
		TotalCount:    totalCount,
		TableURL:      utils.URL("/admin/superuser/logs/table"),
		Include:       "#logs-filter-form",
		ContentTarget: "#logs-table-content",
	}

	if err := compadmin.AdminSuperuserLogsTableContent(logsList, pagination).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) adminSuperuserLogsActionsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Superuser Logs Actions Handler]"
	ctx := r.Context()

	actions, err := s.services.staffLog.GetDistinctActions(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := compadmin.ActionOptions(actions).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
