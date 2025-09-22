//go:build imageprocessing

package cmd

import (
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
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

type convertImagesFlags struct {
	inpath  string
	outpath string
	format  string
}

var flagsConvertImages convertImagesFlags

func init() {
	f := cmdConvertImages.Flags
	f().StringVarP(&flagsConvertImages.inpath, "inpath", "p", "", "Path to the images to process")
	f().StringVarP(&flagsConvertImages.outpath, "outpath", "o", "", "Output path to the images to store")
	f().StringVarP(&flagsConvertImages.format, "format", "f", "webp", "Format of the images to store")
	if err := cmdConvertImages.MarkFlagRequired("inpath"); err != nil {
		panic(errors.Join(errs.ErrCmdRequired, err))
	}
	if err := cmdConvertImages.MarkFlagRequired("outpath"); err != nil {
		panic(errors.Join(errs.ErrCmdRequired, err))
	}
	rootCmd.AddCommand(cmdConvertImages)
}

var cmdConvertImages = &cobra.Command{
	Use:   "convert_images",
	Short: "convert images",
	Run: func(cmd *cobra.Command, args []string) {
		vips.Startup(nil)
		defer vips.Shutdown()

		if _, err := os.Stat(flagsConvertImages.inpath); err != nil || os.IsNotExist(err) {
			panic(errors.Join(errs.ErrCmdRequired, err))
		}
		if _, err := os.Stat(flagsConvertImages.outpath); err != nil || os.IsNotExist(err) {
			if err := os.MkdirAll(flagsConvertImages.outpath, 0755); err != nil {
				panic(errors.Join(errs.ErrCmd, err))
			}
		}

		var ep *vips.ExportParams
		switch flagsConvertImages.format {
		case "png":
			ep = vips.NewDefaultPNGExportParams()
		case "jpg", "jpeg":
			ep = vips.NewDefaultJPEGExportParams()
		case "webp":
			ep = vips.NewDefaultWEBPExportParams()
		default:
			panic("Invalid format: " + flagsConvertImages.format)
		}

		if err := filepath.Walk(flagsConvertImages.inpath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				logs.Log().Info("Skipping directory", zap.String("path", path))
				return nil
			}

			ext := filepath.Ext(info.Name())
			if !utils.IsValidImageExtension(ext){
				logs.Log().Info("Skipping non-image file", zap.String("path", path))
				return nil
			}

			filename := strings.TrimSuffix(info.Name(), ext)
			output := fmt.Sprintf(
				"%s/%s.%s",
				flagsConvertImages.outpath,
				filename,
				flagsConvertImages.format,
			)

			if _, err := os.Stat(output); err == nil {
				logs.Log().Info("Output file already exists, skipping",
					zap.String("input", path),
					zap.String("output", output))
				return nil
			}

			logs.Log().Info("Processing image", zap.String("path", path))

			img, err := vips.NewImageFromFile(path)
			if err != nil {
				logs.Log().Error("Failed to load image", zap.Error(err), zap.String("path", path))
				return nil
			}

			imgbytes, _, err := img.Export(ep)
			if err != nil {
				logs.Log().Error("Failed to export image", zap.Error(err))
				img.Close()
				return err
			}

			if err := os.WriteFile(output, imgbytes, 0644); err != nil {
				logs.Log().Error("Failed to write output file", zap.Error(err))
				img.Close()
				return err
			}

			logs.Log().Info("Successfully converted image",
				zap.String("input", path),
				zap.String("output", output))

			img.Close()
			return nil
		}); err != nil {
			panic(errors.Join(errs.ErrCmd, err))
		}
	},
}
