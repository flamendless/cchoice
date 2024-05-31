package cmd

import (
	"cchoice/api"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var ctxAPI ctx.APIFlags

func init() {
	f := runAPIServerCmd.Flags
	f().StringVarP(&ctxAPI.Address, "address", "a", "", "Address to use")
	f().IntVarP(&ctxAPI.Port, "port", "p", 3000, "Port of the address")

	rootCmd.AddCommand(runAPIServerCmd)
}

var runAPIServerCmd = &cobra.Command{
	Use:   "run_api_server",
	Short: "Run the API server",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Log().Info(
			"Run API Server",
			zap.String("address", ctxAPI.Address),
			zap.Int("port", ctxAPI.Port),
		)

		api.Serve(ctxAPI)
	},
}
