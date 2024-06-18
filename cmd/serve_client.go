package cmd

import (
	"cchoice/client"
	"cchoice/internal/ctx"
	"cchoice/internal/logs"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var ctxClient ctx.ClientFlags

func init() {
	f := serveClientCmd.Flags
	f().BoolVarP(&ctxClient.Secure, "secure", "s", true, "Use secure session")
	f().StringVarP(&ctxClient.Address, "address", "a", "localhost", "Address")
	f().StringVarP(&ctxClient.Port, "port", "p", ":3001", "Port of the address")
	f().StringVarP(&ctxClient.GRPCAddress, "grpc_address", "g", "", "Address for GRPC connection")

	rootCmd.AddCommand(serveClientCmd)
}

var serveClientCmd = &cobra.Command{
	Use:   "serve_client",
	Short: "Run the client server",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Log().Info(
			"Run the client server",
			zap.Bool("secure", ctxClient.Secure),
			zap.String("address", ctxClient.Address),
			zap.String("port", ctxClient.Port),
			zap.String("grpc address", ctxClient.GRPCAddress),
		)
		client.Serve(&ctxClient)
	},
}
