package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var flagPassword string

func init() {
	f := cmdGenBasicAuth.Flags
	f().StringVarP(&flagPassword, "password", "p", "", "Password")

	if err := cmdGenBasicAuth.MarkFlagRequired("password"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(cmdGenBasicAuth)
}

var cmdGenBasicAuth = &cobra.Command{
	Use:   "gen_basic_auth",
	Short: "Generate basic auth",
	Run: func(cmd *cobra.Command, args []string) {
		hash, _ := bcrypt.GenerateFromPassword([]byte(flagPassword), bcrypt.DefaultCost)
		fmt.Println("Hashed password:", string(hash))
	},
}
