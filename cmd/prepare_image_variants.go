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

type prepareImageVariantsFlags struct {
	inpath  string
	outpath string
}

type imageSize struct {
	width  int
	height int
}

var imageSizes = []imageSize{
	{96, 96},
	{256, 256},
	{640, 640},
}

var flagsPrepareImageVariants prepareImageVariantsFlags

func getImageFiles(inpath string) ([]string, error) {
	var imageFiles []string

	err := filepath.Walk(inpath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(info.Name())
		if utils.IsValidImageExtension(ext) {
			imageFiles = append(imageFiles, path)
		} else {
			logs.Log().Info("Skipping non-image file", zap.String("path", path))
		}

		return nil
	})

	return imageFiles, err
}

func processImageForSize(imagePath string, size imageSize, webpPath string, webpExport *vips.ExportParams) error {
	folderName := fmt.Sprintf("%dx%d", size.width, size.height)

	img, err := vips.NewImageFromFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}
	defer img.Close()

	if err := img.Thumbnail(size.width, size.height, vips.InterestingCentre); err != nil {
		return fmt.Errorf("failed to create thumbnail: %w", err)
	}

	imgBytes, _, err := img.Export(webpExport)
	if err != nil {
		return fmt.Errorf("failed to export WebP: %w", err)
	}

	filename := strings.TrimSuffix(filepath.Base(imagePath), filepath.Ext(imagePath))
	outputFile := filepath.Join(webpPath, folderName, filename+".webp")

	if err := os.WriteFile(outputFile, imgBytes, 0644); err != nil {
		return fmt.Errorf("failed to write image variant: %w", err)
	}

	logs.Log().Info("Created image variant",
		zap.String("size", folderName),
		zap.String("input", imagePath),
		zap.String("output", outputFile),
	)

	return nil
}

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

		imageFiles, err := getImageFiles(flagsPrepareImageVariants.inpath)
		if err != nil {
			panic(errors.Join(errs.ErrCmd, fmt.Errorf("failed to get image files: %w", err)))
		}

		if len(imageFiles) == 0 {
			logs.Log().Info("No valid image files found", zap.String("path", flagsPrepareImageVariants.inpath))
			return
		}

		logs.Log().Info("Found image files to process", zap.Int("count", len(imageFiles)))

		webpExport := vips.NewDefaultWEBPExportParams()

		for _, size := range imageSizes {
			folderName := fmt.Sprintf("%dx%d", size.width, size.height)
			logs.Log().Info("Processing size variant", zap.String("size", folderName))

			for _, imagePath := range imageFiles {
				logs.Log().Info("Processing image",
					zap.String("path", imagePath),
					zap.String("size", folderName))

				if err := processImageForSize(imagePath, size, webpPath, webpExport); err != nil {
					logs.Log().Error("Failed to process image variant",
						zap.Error(err),
						zap.String("path", imagePath),
						zap.String("size", folderName))
					continue
				}
			}

			logs.Log().Info("Completed size variant",
				zap.String("size", folderName),
				zap.Int("processed", len(imageFiles)))
		}

		logs.Log().Info("Image variant preparation completed",
			zap.String("brand", brand),
			zap.String("output", brandPath))
	},
}
