package cmd

import (
	"cchoice/internal/ctx"
	"cchoice/internal/logs"
	"cchoice/site"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var ctxSite ctx.SiteFlags

func init() {
	f := serveSiteCmd.Flags
	f().StringVarP(&ctxSite.Port, "port", "p", ":3001", "Port of the address")

	rootCmd.AddCommand(serveSiteCmd)
}

var serveSiteCmd = &cobra.Command{
	Use:   "serve_site",
	Short: "Run the site server",
	Run: func(cmd *cobra.Command, args []string) {
		logs.Log().Info(
			"Run the site server",
			zap.String("port", ctxSite.Port),
		)
		site.Serve(&ctxSite)
	},
}
