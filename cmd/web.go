package cmd

import (
	"cchoice/internal/logs"
	"cchoice/internal/server"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdWeb)
}

var cmdWeb = &cobra.Command{
	Use:   "web",
	Short: "Run the web frontend only (no payment/shipping/mail services)",
	Run: func(cmd *cobra.Command, args []string) {
		serverInstance := server.NewServer()
		done := make(chan bool, 1)

		go gracefulShutdown(serverInstance, done)

		httpServer := serverInstance.HTTPServer

		logs.Log().Info("Serving HTTP (web mode)")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("http server error: %s", err))
		}

		<-done
		logs.Log().Info("Graceful shutdown complete.")
	},
}
