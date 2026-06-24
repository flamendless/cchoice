package services

import (
	"context"
	"encoding/csv"
	"strconv"

	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/logs"

	"go.uber.org/zap"
)

type ProductExportRow struct {
	Brand            string
	Serial           string
	Slug             string
	Status           string
	Category         string
	Subcategory      string
	Name             string
	UnitPriceWithVat string
	SalePriceWithVat string
	SaleStartDate    string
	SaleEndDate      string
	Description      string
	Colours          string
	Sizes            string
	Segmentation     string
	PartNumber       string
	Power            string
	Capacity         string
	Weight           string
	WeightUnit       string
	ScopeOfSupply    string
	StocksIn         string
	StocksQty        string
	ImageURL       string
	ThumbnailURL   string
	CreatedAt      string
	UpdatedAt      string
}

var productExportHeaders = []string{
	"row number",
	"brand",
	"serial number",
	"product name",
	"slug",
	"status",
	"category",
	"subcategory",
	"unit price with vat",
	"sale price with vat",
	"sale start date",
	"sale end date",
	"description",
	"colours",
	"sizes",
	"segmentation",
	"part number",
	"power",
	"capacity",
	"weight",
	"weight unit",
	"scope of supply",
	"stocks in",
	"stocks qty",
	"product image filename or cdn url",
	"product image thumbnail filename or cdn url",
	"created at",
	"updated at",
}

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
) error {
	result := "success"
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
		if err := writer.Write([]string{
			strconv.Itoa(i + 1),
			row.Brand,
			row.Serial,
			row.Name,
			row.Slug,
			row.Status,
			row.Category,
			row.Subcategory,
			row.UnitPriceWithVat,
			row.SalePriceWithVat,
			row.SaleStartDate,
			row.SaleEndDate,
			row.Description,
			row.Colours,
			row.Sizes,
			row.Segmentation,
			row.PartNumber,
			row.Power,
			row.Capacity,
			row.Weight,
			row.WeightUnit,
			row.ScopeOfSupply,
			row.StocksIn,
			row.StocksQty,
			row.ImageURL,
			row.ThumbnailURL,
			row.CreatedAt,
			row.UpdatedAt,
		}); err != nil {
			result = err.Error()
			return err
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
