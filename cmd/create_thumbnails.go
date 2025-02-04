package cmd

import (
	"cchoice/internal/logs"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/davidbyttow/govips/v2/vips"
)

type createThumbnailsFlags struct {
	inpath  string
	outpath string
	width   int
	height  int
}

var createThumbnailFlags createThumbnailsFlags

func init() {
	f := cmdCreateThumbnails.Flags
	f().StringVarP(&createThumbnailFlags.inpath, "inpath", "p", "", "Path to the images to process")
	f().StringVarP(&createThumbnailFlags.outpath, "outpath", "o", "", "Output path to the images to store")
	f().IntVarP(&createThumbnailFlags.width, "width", "", 160, "Width of the output thumbnail")
	f().IntVarP(&createThumbnailFlags.height, "height", "", 160, "Height of the output thumbnail")
	if err := cmdCreateThumbnails.MarkFlagRequired("inpath"); err != nil {
		panic(err)
	}
	if err := cmdCreateThumbnails.MarkFlagRequired("outpath"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(cmdCreateThumbnails)
}

var cmdCreateThumbnails = &cobra.Command{
	Use:   "create_thumbnails",
	Short: "Create thumbnails",
	Run: func(cmd *cobra.Command, args []string) {
		vips.Startup(nil)
		defer vips.Shutdown()

		if _, err := os.Stat(createThumbnailFlags.inpath); err != nil || os.IsNotExist(err) {
			panic(err)
		}
		if _, err := os.Stat(createThumbnailFlags.outpath); err != nil || os.IsNotExist(err) {
			if err := os.MkdirAll(createThumbnailFlags.outpath, 0755); err != nil {
				panic(err)
			}
		}

		if err := filepath.Walk(createThumbnailFlags.inpath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				logs.Log().Info("Skipping...", zap.String("path", path))
				return nil
			}

			img, err := vips.NewImageFromFile(path)
			if err != nil {
				return nil
			}

			if err := img.Thumbnail(
				createThumbnailFlags.width,
				createThumbnailFlags.height,
				vips.InterestingCentre,
			); err != nil {
				return nil
			}

			ep := vips.NewDefaultPNGExportParams()
			imgbytes, _, err := img.Export(ep)
			if err != nil {
				return err
			}

			output := fmt.Sprintf("%s/%s", createThumbnailFlags.outpath, info.Name())
			if err := os.WriteFile(output, imgbytes, 0644); err != nil {
				return err
			}

			return nil
		}); err != nil {
			panic(err)
		}
	},
}
