package cmd

import (
	"cchoice/internal/receipt/scanner/googlevision"
	"fmt"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdTestGVision)
}

var cmdTestGVision = &cobra.Command{
	Use:   "test_gvision",
	Short: "Test Google Vision connection",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Testing Google Vision Connection...")
		fmt.Println("==========================================")
		s := googlevision.MustInit()
		defer s.Close()
		fmt.Println("==========================================")
		fmt.Println("All tests passed! Google Vision is accessible.")
	},
}
