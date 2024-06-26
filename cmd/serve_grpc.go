package cmd

import (
	"cchoice/grpc_server"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var ctxGRPC ctx.GRPCFlags

func init() {
	f := serveGRPCCmd.Flags
	f().StringVarP(&ctxGRPC.Address, "address", "a", ":50051", "Address to use")
	f().BoolVarP(&ctxGRPC.Reflection, "reflection", "r", false, "Allow reflection or not")
	f().BoolVarP(&ctxGRPC.LogPayloadReceived, "log_payload_received", "", false, "Log GRPC payload received")
	f().BoolVarP(&ctxGRPC.RateLimit, "rate_limit", "", true, "Rate limit")
	f().StringVarP(&ctxGRPC.DBPath, "db_path", "", ":memory:", "Path to database")

	rootCmd.AddCommand(serveGRPCCmd)
}

var serveGRPCCmd = &cobra.Command{
	Use:   "serve_grpc",
	Short: "Run the GRPC server",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Log().Info(
			"Run the GRPC Server",
			zap.String("address", ctxGRPC.Address),
			zap.Bool("reflection", ctxGRPC.Reflection),
			zap.Bool("log payload received", ctxGRPC.LogPayloadReceived),
			zap.Bool("rate limit", ctxGRPC.RateLimit),
			zap.String("DB Path", ctxGRPC.DBPath),
		)

		grpc_server.Serve(ctxGRPC)
	},
}
