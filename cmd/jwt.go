package cmd

import (
	"cchoice/internal/auth"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	subcmd      string
	tokenString string
	dbPath      string
)

func init() {
	f := JWTCmd.Flags
	f().StringVarP(&subcmd, "subcmd", "s", "", "subcommand to use")
	f().StringVarP(&tokenString, "token", "t", "", "token to validate")
	f().StringVarP(&dbPath, "db", "", "", "db path")

	rootCmd.AddCommand(JWTCmd)
}

func issue() string {
	issuer, err := auth.NewIssuer()
	if err != nil {
		logs.Log().Error("Unable to create issuer", zap.Error(err))
		panic(1)
	}

	token, err := issuer.IssueToken(enums.AudSystem, "test")
	if err != nil {
		logs.Log().Error("Unable to issue token", zap.Error(err))
		panic(1)
	}

	return token
}

func validate(tokenString string) *jwt.Token {
	ctxDB := ctx.NewDatabaseCtx(dbPath)
	defer ctxDB.Close()

	v, err := auth.NewValidator(ctxDB)
	if err != nil {
		logs.Log().Error("Unable to create validator", zap.Error(err))
		panic(1)
	}

	token, err := v.GetToken(tokenString)
	if err != nil {
		logs.Log().Error("Unable to get validated token", zap.Error(err))
		os.Exit(1)
	}

	return token
}

var JWTCmd = &cobra.Command{
	Use:   "jwt",
	Short: "JWT",
	Run: func(cmd *cobra.Command, args []string) {
		if subcmd == "issue" {
			tokenString := issue()
			logs.Log().Info("issued token", zap.String("token", tokenString))

		} else if subcmd == "validate" {
			token := validate(tokenString)
			logs.Log().Info("validated token", zap.Any("token", token))

		} else if subcmd == "both" {
			tokenString := issue()
			logs.Log().Info("issued token", zap.String("token", tokenString))

			token := validate(tokenString)
			logs.Log().Info("validated token", zap.Any("token", token))

			if !token.Valid {
				panic("Not valid token")
			}

		} else {
			panic("must pass a subcommand: issue, validate, both")
		}
	},
}
