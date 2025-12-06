package cmd

import (
	"cchoice/internal/logs"
	"cchoice/internal/server"
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(cmdAPI)
}

func gracefulShutdown(serverInstance *server.ServerInstance, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logs.Log().Info("shutting down gracefully, press Ctrl+C again to force")

	serverInstance.StopBackgroundJobs()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serverInstance.HTTPServer.Shutdown(ctx); err != nil {
		logs.Log().Error("Server forced to shutdown with error: %v", zap.Error(err))
	}

	logs.Log().Info("Server exiting")

	done <- true
}

var cmdAPI = &cobra.Command{
	Use:   "api",
	Short: "Run the api",
	Run: func(cmd *cobra.Command, args []string) {
		serverInstance := server.NewServer()
		done := make(chan bool, 1)

		serverInstance.StartBackgroundJobs()

		go gracefulShutdown(serverInstance, done)

		httpServer := serverInstance.HTTPServer
		if httpServer.TLSConfig != nil && len(httpServer.TLSConfig.Certificates) > 0 {
			logs.Log().Info("Serving secure HTTP")
			if err := httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				panic(fmt.Sprintf("http server error: %s", err))
			}
		} else {
			logs.Log().Info("Serving HTTP")
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				panic(fmt.Sprintf("http server error: %s", err))
			}
		}

		<-done
		logs.Log().Info("Graceful shutdown complete.")
	},
}
