package cmd

import (
	"cchoice/internal/auth"
	"cchoice/internal/logs"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	subcmd      string
	tokenString string
)

func init() {
	f := JWTCmd.Flags
	f().StringVarP(&subcmd, "subcmd", "s", "", "subcommand to use")
	f().StringVarP(&tokenString, "token", "t", "", "token to validate")

	rootCmd.AddCommand(JWTCmd)
}

var JWTCmd = &cobra.Command{
	Use:   "jwt",
	Short: "JWT",
	Run: func(cmd *cobra.Command, args []string) {
		if subcmd == "issue" {
			issuer, err := auth.NewIssuer()
			if err != nil {
				logs.Log().Error("Unable to create issuer", zap.Error(err))
				panic(1)
			}

			token, err := issuer.IssueToken("test", []string{"test"})
			if err != nil {
				logs.Log().Error("Unable to issue token", zap.Error(err))
				panic(1)
			}

			logs.Log().Info("issued token", zap.String("token", token))

		} else if subcmd == "validate" {
			v, err := auth.NewValidator()
			if err != nil {
				logs.Log().Error("Unable to create validator", zap.Error(err))
				panic(1)
			}

			token, err := v.GetToken(tokenString)
			if err != nil {
				logs.Log().Error("Unable to get validated token", zap.Error(err))
				os.Exit(1)
			}

			logs.Log().Info("validated token", zap.Any("token", token))

		} else {
			panic("must pass a subcommand: issue, validate")
		}
	},
}
