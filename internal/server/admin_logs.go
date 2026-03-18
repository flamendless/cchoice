package server

import (
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
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

	logsList, err := s.services.staffLog.GetAllAsModel(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := compadmin.AdminSuperuserLogsTable(logsList).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
