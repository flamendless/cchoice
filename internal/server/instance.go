package server

import (
	"context"
	"net/http"

	"cchoice/internal/conf"
	"cchoice/internal/logs"
)

type ServerInstance struct {
	HTTPServer    *http.Server
	internal      *Server
	jobRunnerCtx  context.Context
	jobRunnerStop context.CancelFunc
}

func (si *ServerInstance) StartBackgroundJobs() {
	cfg := conf.Conf()
	if cfg.IsProd() && si.internal.mailJobRunner == nil {
		panic("No email job runner initialized")
	}
	if cfg.IsProd() && si.internal.thumbnailJobRunner == nil {
		panic("No thumbnail job runner initialized")
	}

	si.jobRunnerCtx, si.jobRunnerStop = context.WithCancel(context.Background())
	if si.internal.mailJobRunner != nil {
		go si.internal.mailJobRunner.Start(si.jobRunnerCtx)
	}
	if si.internal.thumbnailJobRunner != nil {
		go si.internal.thumbnailJobRunner.Start(si.jobRunnerCtx)
	}
	logs.Log().Info("Background job runners started")
}

func (si *ServerInstance) StopBackgroundJobs() {
	if si.jobRunnerStop != nil {
		si.jobRunnerStop()
		logs.Log().Info("Background job runners stopped")
	}
}
