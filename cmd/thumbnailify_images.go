package cmd

import (
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/davidbyttow/govips/v2/vips"
)

type thumbnailifyImagesFlags struct {
	inpath  string
	outpath string
	format  string
	width   int
	height  int
}

var flagsThumbnailifyImages thumbnailifyImagesFlags

func init() {
	f := cmdThumbnailifyImages.Flags
	f().StringVarP(&flagsThumbnailifyImages.inpath, "inpath", "p", "", "Path to the images to process")
	f().StringVarP(&flagsThumbnailifyImages.outpath, "outpath", "o", "", "Output path to the images to store")
	f().StringVarP(&flagsThumbnailifyImages.format, "format", "f", "png", "Format of the images to store")
	f().IntVarP(&flagsThumbnailifyImages.width, "width", "", 160, "Width of the output thumbnail")
	f().IntVarP(&flagsThumbnailifyImages.height, "height", "", 160, "Height of the output thumbnail")
	if err := cmdThumbnailifyImages.MarkFlagRequired("inpath"); err != nil {
		panic(errors.Join(errs.ErrCmdRequired, err))
	}
	if err := cmdThumbnailifyImages.MarkFlagRequired("outpath"); err != nil {
		panic(errors.Join(errs.ErrCmdRequired, err))
	}
	rootCmd.AddCommand(cmdThumbnailifyImages)
}

var cmdThumbnailifyImages = &cobra.Command{
	Use:   "thumbnailify_images",
	Short: "thumbnailify images",
	Run: func(cmd *cobra.Command, args []string) {
		vips.Startup(nil)
		defer vips.Shutdown()

		if _, err := os.Stat(flagsThumbnailifyImages.inpath); err != nil || os.IsNotExist(err) {
			panic(errors.Join(errs.ErrCmdRequired, err))
		}
		if _, err := os.Stat(flagsThumbnailifyImages.outpath); err != nil || os.IsNotExist(err) {
			if err := os.MkdirAll(flagsThumbnailifyImages.outpath, 0755); err != nil {
				panic(errors.Join(errs.ErrCmd, err))
			}
		}

		var ep *vips.ExportParams
		switch flagsThumbnailifyImages.format {
		case "png":
			ep = vips.NewDefaultPNGExportParams()
		case "jpg", "jpeg":
			ep = vips.NewDefaultJPEGExportParams()
		case "webp":
			ep = vips.NewDefaultWEBPExportParams()
		default:
			panic("Invalid format: " + flagsThumbnailifyImages.format)
		}

		if err := filepath.Walk(flagsThumbnailifyImages.inpath, func(path string, info fs.FileInfo, err error) error {
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
				flagsThumbnailifyImages.width,
				flagsThumbnailifyImages.height,
				vips.InterestingCentre,
			); err != nil {
				return nil
			}

			imgbytes, _, err := img.Export(ep)
			if err != nil {
				return err
			}

			filename := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			output := fmt.Sprintf(
				"%s/%s_%dx%d.%s",
				flagsThumbnailifyImages.outpath,
				filename,
				flagsThumbnailifyImages.width,
				flagsThumbnailifyImages.height,
				flagsThumbnailifyImages.format,
			)
			if err := os.WriteFile(output, imgbytes, 0644); err != nil {
				return err
			}

			return nil
		}); err != nil {
			panic(errors.Join(errs.ErrCmd, err))
		}
	},
}
