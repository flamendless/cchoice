//TODO: (Brandon)
//this should be a test

package cmd

import (
	"cchoice/internal/auth"
	"cchoice/internal/ctx"
	"cchoice/internal/enums"
	"cchoice/internal/logs"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	subcmd      string
	tokenString string
	audString   string
	dbPath      string
	username    string
	tokenOnly   bool
)

func init() {
	f := JWTCmd.Flags
	f().StringVarP(&subcmd, "subcmd", "s", "", "subcommand to use")
	f().StringVarP(&tokenString, "token", "t", "", "token to validate")
	f().StringVarP(&dbPath, "db", "", "", "db path")
	f().StringVarP(&audString, "aud", "a", "SYSTEM", "AUD")
	f().StringVarP(&username, "username", "u", "client@cchoice.com", "username")
	f().BoolVarP(&tokenOnly, "token_only", "o", false, "whether to output token in shell")

	rootCmd.AddCommand(JWTCmd)
}

func issue() string {
	issuer, err := auth.NewIssuer()
	if err != nil {
		logs.Log().Error("Unable to create issuer", zap.Error(err))
		panic(1)
	}

	aud := enums.ParseAudEnum(audString)
	if aud == enums.AUD_UNDEFINED {
		panic("Invalid AUD string")
	}

	token, err := issuer.IssueToken(aud, username)
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

	expectedAUD := enums.ParseAudEnum(audString)
	res, err := v.GetToken(expectedAUD, tokenString)
	if err != nil {
		logs.Log().Error("Unable to get validated token", zap.Error(err))
		os.Exit(1)
	}

	return res.Token
}

var JWTCmd = &cobra.Command{
	Use:   "jwt",
	Short: "JWT",
	Run: func(cmd *cobra.Command, args []string) {
		switch subcmd {
		case "issue":
			tokenString := issue()
			logs.Log().Info(
				"issued token",
				zap.String("aud", audString),
				zap.String("token", tokenString),
			)
			if tokenOnly {
				fmt.Println(tokenString)
			}
		case "validate":
			token := validate(tokenString)
			logs.Log().Info("validated token", zap.Any("token", token))
		case "both":
			tokenString := issue()
			logs.Log().Info(
				"issued token",
				zap.String("aud", audString),
				zap.String("token", tokenString),
			)

			token := validate(tokenString)
			logs.Log().Info("validated token", zap.Any("token", token))

			if !token.Valid {
				panic("Not valid token")
			}
		default:
			panic("must pass a subcommand: issue, validate, both")
		}
	},
}
