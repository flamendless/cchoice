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
	if cfg.IsProd() {
		if si.internal.mailJobRunner == nil {
			panic("No email job runner initialized")
		}
		if si.internal.thumbnailJobRunner == nil {
			panic("No thumbnail job runner initialized")
		}
		if si.internal.invoiceJobRunner == nil {
			panic("No invoice job runner initialized")
		}
	}

	si.jobRunnerCtx, si.jobRunnerStop = context.WithCancel(context.Background())
	if si.internal.mailJobRunner != nil {
		go si.internal.mailJobRunner.Start(si.jobRunnerCtx)
	}
	if si.internal.thumbnailJobRunner != nil {
		go si.internal.thumbnailJobRunner.Start(si.jobRunnerCtx)
	}
	if si.internal.invoiceJobRunner != nil {
		go si.internal.invoiceJobRunner.Start(si.jobRunnerCtx)
	}
	logs.Log().Info("Background job runners started")
}

func (si *ServerInstance) StopBackgroundJobs() {
	if si.jobRunnerStop != nil {
		si.jobRunnerStop()
		logs.Log().Info("Background job runners stopped")
	}
	if si.internal.rateLimiter != nil {
		si.internal.rateLimiter.Stop()
		logs.Log().Info("Rate limiter stopped")
	}
}
