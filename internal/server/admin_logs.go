package server

import (
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/logs"

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

	var staffID int64
	if staffIDStr != "" {
		decoded := s.encoder.Decode(staffIDStr)
		if decoded != encode.INVALID {
			staffID = decoded
		}
	}

	module := enums.ParseModuleToEnum(moduleStr)

	logsList, err := s.services.staffLog.GetFilteredAsModel(ctx, staffID, action, module)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := compadmin.AdminSuperuserLogsTable(logsList).Render(ctx, w); err != nil {
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


