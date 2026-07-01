package server

import (
	"net/http"
	"strconv"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
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

	staffIDStr := r.URL.Query().Get("staff-id")
	action := r.URL.Query().Get("action")
	moduleStr := r.URL.Query().Get("module")

	page := 1
	if paramPage := r.URL.Query().Get("page"); paramPage != "" {
		if parsed, err := strconv.Atoi(paramPage); err == nil && parsed > 0 {
			page = parsed
		}
	}

	var staffID int64
	if staffIDStr != "" {
		decoded := s.encoder.Decode(staffIDStr)
		if decoded != encode.INVALID {
			staffID = decoded
		}
	}

	module := enums.ParseModuleToEnum(moduleStr)

	logsList, totalCount, page, err := s.services.staffLog.GetFilteredAsModelPaginated(
		ctx,
		staffID,
		action,
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
