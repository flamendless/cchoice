package cmd

import (
	"cchoice/internal/logs"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/davidbyttow/govips/v2/vips"
)

type processImagesFlags struct {
	inpath  string
	outpath string
	format  string
	width   int
	height  int
}

var flagsProcessImages processImagesFlags

func init() {
	f := cmdProcessImages.Flags
	f().StringVarP(&flagsProcessImages.inpath, "inpath", "p", "", "Path to the images to process")
	f().StringVarP(&flagsProcessImages.outpath, "outpath", "o", "", "Output path to the images to store")
	f().StringVarP(&flagsProcessImages.format, "format", "f", "png", "Format of the images to store")
	f().IntVarP(&flagsProcessImages.width, "width", "", 160, "Width of the output thumbnail")
	f().IntVarP(&flagsProcessImages.height, "height", "", 160, "Height of the output thumbnail")
	if err := cmdProcessImages.MarkFlagRequired("inpath"); err != nil {
		panic(err)
	}
	if err := cmdProcessImages.MarkFlagRequired("outpath"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(cmdProcessImages)
}

var cmdProcessImages = &cobra.Command{
	Use:   "process_images",
	Short: "Process images",
	Run: func(cmd *cobra.Command, args []string) {
		vips.Startup(nil)
		defer vips.Shutdown()

		if _, err := os.Stat(flagsProcessImages.inpath); err != nil || os.IsNotExist(err) {
			panic(err)
		}
		if _, err := os.Stat(flagsProcessImages.outpath); err != nil || os.IsNotExist(err) {
			if err := os.MkdirAll(flagsProcessImages.outpath, 0755); err != nil {
				panic(err)
			}
		}

		if err := filepath.Walk(flagsProcessImages.inpath, func(path string, info fs.FileInfo, err error) error {
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
				flagsProcessImages.width,
				flagsProcessImages.height,
				vips.InterestingCentre,
			); err != nil {
				return nil
			}

			var ep *vips.ExportParams
			switch flagsProcessImages.format {
			case "png":
				ep = vips.NewDefaultPNGExportParams()
			case "jpg", "jpeg":
				ep = vips.NewDefaultJPEGExportParams()
			case "webp":
				ep = vips.NewDefaultWEBPExportParams()
			default:
				panic("Invalid format: " + flagsProcessImages.format)
			}
			imgbytes, _, err := img.Export(ep)
			if err != nil {
				return err
			}

			filename := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			output := fmt.Sprintf(
				"%s/%s_%dx%d.%s",
				flagsProcessImages.outpath,
				filename,
				flagsProcessImages.width,
				flagsProcessImages.height,
				flagsProcessImages.format,
			)
			if err := os.WriteFile(output, imgbytes, 0644); err != nil {
				return err
			}

			return nil
		}); err != nil {
			panic(err)
		}
	},
}
