package cmd

import (
	"cchoice/http_server"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var ctxAPI ctx.APIFlags

func init() {
	f := serveHTTPCmd.Flags
	f().StringVarP(&ctxAPI.Address, "address", "a", "", "Address to use")
	f().IntVarP(&ctxAPI.Port, "port", "p", 3000, "Port of the address")

	rootCmd.AddCommand(serveHTTPCmd)
}

var serveHTTPCmd = &cobra.Command{
	Use:   "serve_http",
	Short: "Run the HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Log().Info(
			"Run the HTTP Server",
			zap.String("address", ctxAPI.Address),
			zap.Int("port", ctxAPI.Port),
		)

		http_server.Serve(ctxAPI)
	},
}
