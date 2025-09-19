//go:build imageprocessing

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

type prepareImageVariantsFlags struct {
	inpath  string
	outpath string
}

var imageSizes = []struct {
	width  int
	height int
}{
	{96, 96},
	{256, 256},
	{640, 640},
}

var flagsPrepareImageVariants prepareImageVariantsFlags

func init() {
	f := cmdPrepareImageVariants.Flags
	f().StringVarP(&flagsPrepareImageVariants.inpath, "inpath", "p", "", "Path to the images to process (brand will be determined from last folder)")
	f().StringVarP(&flagsPrepareImageVariants.outpath, "outpath", "o", "", "Output base path (e.g., /static/images/product_images/)")
	if err := cmdPrepareImageVariants.MarkFlagRequired("inpath"); err != nil {
		panic(errors.Join(errs.ErrCmdRequired, err))
	}
	if err := cmdPrepareImageVariants.MarkFlagRequired("outpath"); err != nil {
		panic(errors.Join(errs.ErrCmdRequired, err))
	}
	rootCmd.AddCommand(cmdPrepareImageVariants)
}

var cmdPrepareImageVariants = &cobra.Command{
	Use:   "prepare_image_variants",
	Short: "prepare image variants with multiple sizes and organize by brand",
	Run: func(cmd *cobra.Command, args []string) {
		vips.Startup(nil)
		defer vips.Shutdown()

		if _, err := os.Stat(flagsPrepareImageVariants.inpath); err != nil || os.IsNotExist(err) {
			panic(errors.Join(errs.ErrCmdRequired, err))
		}

		brand := filepath.Base(flagsPrepareImageVariants.inpath)
		if brand == "." || brand == "/" {
			panic(errors.Join(errs.ErrCmdRequired, fmt.Errorf("cannot determine brand from input path: %s", flagsPrepareImageVariants.inpath)))
		}

		logs.Log().Info("Detected brand from input path", zap.String("brand", brand), zap.String("inpath", flagsPrepareImageVariants.inpath))

		brandPath := filepath.Join(flagsPrepareImageVariants.outpath, brand)
		originalPath := filepath.Join(brandPath, "original")
		webpPath := filepath.Join(brandPath, "webp")

		dirs := []string{originalPath, webpPath}
		for _, size := range imageSizes {
			folderName := fmt.Sprintf("%dx%d", size.width, size.height)
			dirs = append(dirs, filepath.Join(webpPath, folderName))
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to create directory %s: %w", dir, err)))
			}
		}

		webpExport := vips.NewDefaultWEBPExportParams()
		validExts := []string{".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp"}

		if err := filepath.Walk(flagsPrepareImageVariants.inpath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(info.Name()))
			isValid := false
			for _, validExt := range validExts {
				if ext == validExt {
					isValid = true
					break
				}
			}
			if !isValid {
				logs.Log().Info("Skipping non-image file", zap.String("path", path))
				return nil
			}

			logs.Log().Info("Processing image", zap.String("path", path))

			img, err := vips.NewImageFromFile(path)
			if err != nil {
				logs.Log().Error("Failed to load image", zap.Error(err), zap.String("path", path))
				return nil
			}

			filename := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			for _, size := range imageSizes {
				folderName := fmt.Sprintf("%dx%d", size.width, size.height)

				imgCopy, err := vips.NewImageFromFile(path)
				if err != nil {
					logs.Log().Error("Failed to load image copy", zap.Error(err))
					continue
				}

				if err := imgCopy.Thumbnail(size.width, size.height, vips.InterestingCentre); err != nil {
					logs.Log().Error("Failed to create image variant", zap.Error(err))
					imgCopy.Close()
					continue
				}

				imgBytes, _, err := imgCopy.Export(webpExport)
				if err != nil {
					logs.Log().Error("Failed to export WebP", zap.Error(err))
					imgCopy.Close()
					continue
				}

				outputFile := filepath.Join(webpPath, folderName, filename+".webp")
				if err := os.WriteFile(outputFile, imgBytes, 0644); err != nil {
					logs.Log().Error("Failed to write image variant", zap.Error(err))
					imgCopy.Close()
					continue
				}

				logs.Log().Info("Created image variant",
					zap.String("size", folderName),
					zap.String("output", outputFile),
				)

				imgCopy.Close()
			}

			img.Close()
			return nil
		}); err != nil {
			panic(errors.Join(errs.ErrCmd, err))
		}

		logs.Log().Info("Image variant preparation completed",
			zap.String("brand", brand),
			zap.String("output", brandPath))
	},
}
