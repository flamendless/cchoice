package server

import (
	"context"
	"net/http"

	"cchoice/internal/logs"
)

type ServerInstance struct {
	HTTPServer    *http.Server
	internal      *Server
	jobRunnerCtx  context.Context
	jobRunnerStop context.CancelFunc
}

func (si *ServerInstance) StartBackgroundJobs() {
	si.jobRunnerCtx, si.jobRunnerStop = context.WithCancel(context.Background())
	go si.internal.emailJobRunner.Start(si.jobRunnerCtx)
	logs.Log().Info("Background job runners started")
}

func (si *ServerInstance) StopBackgroundJobs() {
	if si.jobRunnerStop != nil {
		si.jobRunnerStop()
		logs.Log().Info("Background job runners stopped")
	}
}
