package cmd

import (
	"cchoice/internal/encode"
	"cchoice/internal/encode/sqids"
	"cchoice/internal/errs"
	"fmt"

	"github.com/spf13/cobra"
)

var flagIntID int64
var flagStrID string
var flagMethod string

func init() {
	f := cmdEncode.Flags
	f().Int64VarP(&flagIntID, "int_id", "i", 0, "Database ID (int)")
	f().StringVarP(&flagStrID, "str_id", "s", "", "Database ID (string)")
	f().StringVarP(&flagMethod, "method", "x", "sqids", "Method (sqids)")

	rootCmd.AddCommand(cmdEncode)
}

var cmdEncode = &cobra.Command{
	Use:   "encode",
	Short: "encode database id (int) <-> string",
	Run: func(cmd *cobra.Command, args []string) {
		var encode encode.IEncode
		switch flagMethod {
		case "sqids":
			encode = sqids.MustSqids()
		default:
			panic(errs.ErrCmd)
		}

		if flagIntID != 0 {
			fmt.Printf("%s\n", encode.Encode(flagIntID))
		}

		if flagStrID != "" {
			fmt.Printf("%d\n", encode.Decode(flagStrID))
		}
	},
}
