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
	rootCmd.AddCommand(apiCmd)
}

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logs.Log().Info("shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		logs.Log().Error("Server forced to shutdown with error: %v", zap.Error(err))
	}

	logs.Log().Info("Server exiting")

	done <- true
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Run the api",
	Run: func(cmd *cobra.Command, args []string) {
		server := server.NewServer()
		done := make(chan bool, 1)
		go gracefulShutdown(server, done)

		if server.TLSConfig != nil {
			logs.Log().Info("Serving secure HTTP")
			if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				panic(fmt.Sprintf("http server error: %s", err))
			}
		} else {
			logs.Log().Info("Serving HTTP")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				panic(fmt.Sprintf("http server error: %s", err))
			}
		}

		<-done
		logs.Log().Info("Graceful shutdown complete.")
	},
}
