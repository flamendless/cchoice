package cmd

import (
	"cchoice/internal/enums"
	"cchoice/internal/storage/linode"
	"context"
	"fmt"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cmdTestLinode)
}

var cmdTestLinode = &cobra.Command{
	Use:   "test_linode",
	Short: "test Linode Object Storage connection",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Testing Linode Object Storage Connection...")
		fmt.Println("==========================================")

		client, err := linode.NewClientFromConfigWithBucket(enums.LINODE_BUCKET_PUBLIC)
		if err != nil {
			panic(err)
		}

		ctx := context.Background()
		if err := client.HeadBucket(ctx); err != nil {
			panic(err)
		}

		fmt.Println("\nTesting object listing in images/ folder...")
		objects, err := client.ListObjects(ctx, "images/", 5)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Successfully listed objects\n")
		if len(objects) > 0 {
			fmt.Printf("  Found %d object(s) (showing up to 5):\n", len(objects))
			for i, obj := range objects {
				if i >= 5 {
					break
				}
				fmt.Printf("    - %s (size: %d bytes)\n", obj.Key, obj.Size)
			}
		} else {
			fmt.Println("  Bucket is empty or no objects found with the specified prefix")
		}

		fmt.Println("\nAll tests passed! Linode Object Storage is accessible.")
	},
}
