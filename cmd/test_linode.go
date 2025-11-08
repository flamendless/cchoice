package cmd

import (
	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

		cfg := conf.Conf()

		if cfg.StorageProvider != "linode" || cfg.Linode.Endpoint == "" || cfg.Linode.AccessKey == "" || cfg.Linode.SecretKey == "" || cfg.Linode.Bucket == "" {
			panic(errs.ErrEnvVarRequired)
		}

		ctx := context.Background()

		awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(cfg.Linode.Region),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.Linode.AccessKey,
				cfg.Linode.SecretKey,
				"",
			)),
		)
		if err != nil {
			panic(err)
		}

		endpoint := cfg.Linode.Endpoint
		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			endpoint = "https://" + endpoint
		}

		client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})

		_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(cfg.Linode.Bucket),
		})
		if err != nil {
			panic(err)
		}

		fmt.Println("\nTesting object listing...")
		listInput := &s3.ListObjectsV2Input{
			Bucket:  aws.String(cfg.Linode.Bucket),
			MaxKeys: aws.Int32(5),
		}
		if cfg.Linode.BasePrefix != "" {
			prefix := cfg.Linode.BasePrefix
			if prefix[0] == '/' {
				prefix = prefix[1:]
			}
			listInput.Prefix = aws.String(prefix)
		}

		result, err := client.ListObjectsV2(ctx, listInput)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Successfully listed objects\n")
		if result.KeyCount != nil && *result.KeyCount > 0 {
			fmt.Printf("  Found %d object(s) (showing up to 5):\n", *result.KeyCount)
			for i, obj := range result.Contents {
				if i >= 5 {
					break
				}
				fmt.Printf("    - %s (size: %d bytes)\n", *obj.Key, *obj.Size)
			}
		} else {
			fmt.Println("  Bucket is empty or no objects found with the specified prefix")
		}

		fmt.Println("\nAll tests passed! Linode Object Storage is accessible.")
	},
}
