package server

import (
	"net/http"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
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

	logsData, err := s.services.staffLogs.GetAll(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logsList := make([]models.StaffLog, 0, len(logsData))
	for _, l := range logsData {
		logsList = append(logsList, models.StaffLog{
			ID:         s.encoder.Encode(l.ID),
			StaffID:    s.encoder.Encode(l.StaffID),
			CreatedAt:  l.CreatedAt,
			Action:     l.Action,
			Module:     l.Module,
			Result:     l.Result,
			FirstName:  l.FirstName.String,
			MiddleName: l.MiddleName.String,
			LastName:   l.LastName.String,
		})
	}

	if err := compadmin.AdminSuperuserLogsTable(logsList).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
