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
	f().StringVarP(&ctxGRPC.Port, "port", "p", ":50051", "Port of the address")
	f().BoolVarP(&ctxGRPC.Reflection, "reflection", "r", false, "Allow reflection or not")

	rootCmd.AddCommand(serveGRPCCmd)
}

var serveGRPCCmd = &cobra.Command{
	Use:   "serve_grpc",
	Short: "Run the GRPC server",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Log().Info(
			"Run the GRPC Server",
			zap.String("port", ctxGRPC.Port),
			zap.Bool("reflection", ctxGRPC.Reflection),
		)

		grpc_server.Serve(ctxGRPC)
	},
}
