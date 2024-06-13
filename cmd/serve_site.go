package cmd

import (
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	"cchoice/site/components"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var ctxSite ctx.SiteFlags

func init() {
	f := serveSiteCmd.Flags
	f().BoolVarP(&ctxSite.Secure, "secure", "s", true, "Use secure session")
	f().StringVarP(&ctxSite.Address, "address", "a", "localhost", "Address")
	f().StringVarP(&ctxSite.Port, "port", "p", ":3001", "Port of the address")

	rootCmd.AddCommand(serveSiteCmd)
}

var serveSiteCmd = &cobra.Command{
	Use:   "serve_site",
	Short: "Run the site server",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Log().Info(
			"Run the site server",
			zap.String("address", ctxSite.Address),
			zap.String("port", ctxSite.Port),
		)
		components.Serve(&ctxSite)
	},
}
