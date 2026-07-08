package services

import (
	"context"
	"encoding/csv"
	"fmt"

	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/logs"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type ExportService struct {
	product  *ProductService
	staffLog *StaffLogsService
}

func NewExportService(product *ProductService, staffLog *StaffLogsService) *ExportService {
	if product == nil {
		panic("ProductService is required")
	}
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	return &ExportService{
		product:  product,
		staffLog: staffLog,
	}
}

func (s *ExportService) StreamProductsCSV(
	ctx context.Context,
	writer *csv.Writer,
	brand string,
	status enums.ProductStatus,
	sortColumn enums.ProductExportSortColumn,
	sortDirection enums.ProductExportSortDirection,
	adminStaffID string,
	filename string,
) error {
	result := fmt.Sprintf("success. filename '%s'", filename)
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			adminStaffID,
			constants.ActionExport,
			constants.ModuleProductsExportCSV,
			result,
			nil,
		); err != nil {
			logs.LogCtx(ctx).Error("[ExportService] failed to log products csv export", zap.Error(err))
		}
	}()

	rows, err := s.product.GetForExportAdmin(ctx, brand, status, sortColumn, sortDirection)
	if err != nil {
		result = err.Error()
		return err
	}

	if err := writer.Write(productExportHeaders); err != nil {
		result = err.Error()
		return err
	}

	for i, row := range rows {
		if err := writer.Write(productExportRowToStrings(row, i+1)); err != nil {
			result = err.Error()
			return err
		}
	}

	return nil
}

func (s *ExportService) StreamProductsXLSX(
	ctx context.Context,
	file *excelize.File,
	brand string,
	status enums.ProductStatus,
	sortColumn enums.ProductExportSortColumn,
	sortDirection enums.ProductExportSortDirection,
	adminStaffID string,
	filename string,
) error {
	result := fmt.Sprintf("success. filename '%s'", filename)
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			adminStaffID,
			constants.ActionExport,
			constants.ModuleProductsExportXLSX,
			result,
			nil,
		); err != nil {
			logs.LogCtx(ctx).Error("[ExportService] failed to log products xlsx export", zap.Error(err))
		}
	}()

	rows, err := s.product.GetForExportAdmin(ctx, brand, status, sortColumn, sortDirection)
	if err != nil {
		result = err.Error()
		return err
	}

	const sheet = "Products"
	if _, err := file.NewSheet(sheet); err != nil {
		result = err.Error()
		return err
	}
	if err := file.DeleteSheet("Sheet1"); err != nil {
		result = err.Error()
		return err
	}

	for colIdx, header := range productExportHeaders {
		cell, err := excelize.CoordinatesToCellName(colIdx+1, 1)
		if err != nil {
			result = err.Error()
			return err
		}
		if err := file.SetCellValue(sheet, cell, header); err != nil {
			result = err.Error()
			return err
		}
	}

	for rowIdx, row := range rows {
		values := productExportRowToStrings(row, rowIdx+1)
		for colIdx, value := range values {
			cell, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			if err != nil {
				result = err.Error()
				return err
			}
			if err := file.SetCellValue(sheet, cell, value); err != nil {
				result = err.Error()
				return err
			}
		}
	}

	return nil
}

func (s *ExportService) ID() string {
	return "Export"
}

func (s *ExportService) Log() {
	logs.Log().Info("[ExportService] Loaded")
}

var _ IService = (*ExportService)(nil)
