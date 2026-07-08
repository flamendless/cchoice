package services

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"cchoice/internal/conf"
	"cchoice/internal/enums"
	"cchoice/internal/logs"

	"github.com/xuri/excelize/v2"
)

type ProductBulkImportService struct {
	product  *ProductService
	staffLog *StaffLogsService
}

func NewProductBulkImportService(product *ProductService, staffLog *StaffLogsService) *ProductBulkImportService {
	if product == nil {
		panic("ProductService is required")
	}
	if staffLog == nil {
		panic("StaffLogsService is required")
	}
	return &ProductBulkImportService{
		product:  product,
		staffLog: staffLog,
	}
}

func (s *ProductBulkImportService) ImportFromReader(
	ctx context.Context,
	staffID string,
	filename string,
	reader io.Reader,
) (*BulkImportResult, error) {
	preview, sessionData, err := s.PreviewFromReader(ctx, filename, reader)
	if err != nil {
		return nil, err
	}

	lines := make([]int, 0, preview.SelectableCount)
	for _, row := range preview.Rows {
		if row.Selectable {
			lines = append(lines, row.Line)
		}
	}

	return s.ApplySelected(ctx, staffID, sessionData, lines)
}

func (s *ProductBulkImportService) upsertRow(
	ctx context.Context,
	staffID string,
	row map[string]string,
	headerMap map[string]int,
) (string, error) {
	serial := cellValue(row, colSerialNumber)
	existing, err := s.product.GetBySerial(ctx, serial)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	if errors.Is(err, sql.ErrNoRows) {
		if err := validateCreateRowValues(row); err != nil {
			return "", err
		}
		input, buildErr := s.buildCreateInput(ctx, row)
		if buildErr != nil {
			return "", buildErr
		}
		product, createErr := s.product.Create(ctx, staffID, input)
		if createErr != nil {
			return "", createErr
		}

		statusStr := cellValue(row, colStatus)
		if statusStr != "" {
			status := enums.ParseProductStatusToEnum(statusStr)
			if status == enums.PRODUCT_STATUS_UNDEFINED {
				return "", fmt.Errorf("invalid status: %s", statusStr)
			}
			if status != enums.PRODUCT_STATUS_DRAFT {
				if err := s.product.UpdateStatus(ctx, s.product.EncodeID(product.ID), status); err != nil {
					return "", err
				}
			}
		}

		return "created", nil
	}

	productID := s.product.EncodeID(existing.ID)
	existingEdit, err := s.product.GetByIDForEdit(ctx, productID)
	if err != nil {
		return "", err
	}

	input, err := s.buildUpdateInput(ctx, row, existingEdit, headerMap)
	if err != nil {
		return "", err
	}
	input.ProductID = productID

	if err := s.product.Update(ctx, staffID, input); err != nil {
		return "", err
	}

	return "updated", nil
}

func (s *ProductBulkImportService) buildCreateInput(
	ctx context.Context,
	row map[string]string,
) (CreateProductInput, error) {
	brandID, err := s.resolveBrandID(ctx, cellValue(row, colBrand))
	if err != nil {
		return CreateProductInput{}, err
	}

	prices, err := parseRowPrices(row)
	if err != nil {
		return CreateProductInput{}, err
	}

	specs, err := parseRowSpecs(row)
	if err != nil {
		return CreateProductInput{}, err
	}

	stocksIn, stocksQty, err := parseRowStocks(row)
	if err != nil {
		return CreateProductInput{}, err
	}

	saleStartDate, saleEndDate, err := parseRowSaleDates(row, prices.salePriceWithVat > 0)
	if err != nil {
		return CreateProductInput{}, err
	}

	externalLinks, err := validateRowExternalLinks(row)
	if err != nil {
		return CreateProductInput{}, err
	}

	return CreateProductInput{
		Serial:              cellValue(row, colSerialNumber),
		Name:                cellValue(row, colProductName),
		Description:         cellValue(row, colDescription),
		BrandID:             brandID,
		Category:            cellValue(row, colCategory),
		Subcategory:         cellValue(row, colSubcategory),
		Specs:               specs,
		UnitPriceWithoutVat: prices.unitPriceWithoutVat,
		UnitPriceWithVat:    prices.unitPriceWithVat,
		SalePriceWithoutVat: prices.salePriceWithoutVat,
		SalePriceWithVat:    prices.salePriceWithVat,
		SaleStartDate:       saleStartDate,
		SaleEndDate:         saleEndDate,
		StocksIn:            stocksIn,
		Stocks:              stocksQty,
		ExternalLinks:       externalLinks,
	}, nil
}

func (s *ProductBulkImportService) buildUpdateInput(
	ctx context.Context,
	row map[string]string,
	existing *ProductForEdit,
	headerMap map[string]int,
) (UpdateProductInput, error) {
	brandName := mergeImportString(row, colBrand, existing.BrandName, importBlankBehavior(colBrand))
	brandID, err := s.resolveBrandID(ctx, brandName)
	if err != nil {
		return UpdateProductInput{}, err
	}

	weightUnit := mergeImportString(row, colWeightUnit, existing.Specs.WeightUnit, importBlankBehavior(colWeightUnit))
	weightUnit, err = normalizeImportWeightUnit(weightUnit)
	if err != nil {
		return UpdateProductInput{}, err
	}

	specs := ProductSpecsInput{
		Colours:       mergeImportString(row, colColours, existing.Specs.Colours, importBlankBehavior(colColours)),
		Sizes:         mergeImportString(row, colSizes, existing.Specs.Sizes, importBlankBehavior(colSizes)),
		Segmentation:  mergeImportString(row, colSegmentation, existing.Specs.Segmentation, importBlankBehavior(colSegmentation)),
		PartNumber:    mergeImportString(row, colPartNumber, existing.Specs.PartNumber, importBlankBehavior(colPartNumber)),
		Power:         mergeImportString(row, colPower, existing.Specs.Power, importBlankBehavior(colPower)),
		Capacity:      mergeImportString(row, colCapacity, existing.Specs.Capacity, importBlankBehavior(colCapacity)),
		Weight:        mergeImportString(row, colWeight, existing.Specs.Weight, importBlankBehavior(colWeight)),
		WeightUnit:    weightUnit,
		ScopeOfSupply: mergeImportString(row, colScopeOfSupply, existing.Specs.ScopeOfSupply, importBlankBehavior(colScopeOfSupply)),
	}

	unitPriceWithVat := existing.UnitPriceWithVat / 100
	if cellProvided(row, colUnitPriceWithVat) {
		parsed, parseErr := parseExportPrice(cellValue(row, colUnitPriceWithVat))
		if parseErr != nil {
			return UpdateProductInput{}, parseErr
		}
		unitPriceWithVat = parsed
	}

	vatPercentage, err := strconv.ParseFloat(conf.Conf().Settings.VATPercentage, 64)
	if err != nil {
		vatPercentage = 0
	}
	unitPriceWithoutVat := int64(math.Round(float64(unitPriceWithVat) / (1 + vatPercentage/100)))

	salePriceWithVat := existing.SalePriceWithVat / 100
	salePriceWithoutVat := int64(0)
	if salePriceWithVat > 0 {
		salePriceWithoutVat = int64(math.Round(float64(salePriceWithVat) / (1 + vatPercentage/100)))
	}
	saleStartDate := existing.SaleStartDate
	saleEndDate := existing.SaleEndDate

	if cellProvided(row, colSalePriceWithVat) {
		parsed, parseErr := parseExportPrice(cellValue(row, colSalePriceWithVat))
		if parseErr != nil {
			return UpdateProductInput{}, parseErr
		}
		salePriceWithVat = parsed
		salePriceWithoutVat = int64(math.Round(float64(salePriceWithVat) / (1 + vatPercentage/100)))

		if cellProvided(row, colSaleStartDate) {
			saleStartDate, err = parseExportDate(cellValue(row, colSaleStartDate))
			if err != nil {
				return UpdateProductInput{}, err
			}
		}
		if cellProvided(row, colSaleEndDate) {
			saleEndDate, err = parseExportDate(cellValue(row, colSaleEndDate))
			if err != nil {
				return UpdateProductInput{}, err
			}
		}
		if salePriceWithVat > 0 && (saleStartDate == "" || saleEndDate == "") {
			return UpdateProductInput{}, errors.New("sale start and end dates are required when sale price is set")
		}
	}

	stocksIn := existing.StocksIn
	stocksQty := existing.Stocks
	if cellProvided(row, colStocksIn) {
		parsed := enums.ParseStocksInToEnum(cellValue(row, colStocksIn))
		if parsed == enums.STOCKS_IN_UNDEFINED {
			return UpdateProductInput{}, fmt.Errorf("invalid stocks in: %s", cellValue(row, colStocksIn))
		}
		stocksIn = parsed
	}
	if cellProvided(row, colStocksQty) {
		parsed, parseErr := strconv.ParseInt(cellValue(row, colStocksQty), 10, 64)
		if parseErr != nil {
			return UpdateProductInput{}, fmt.Errorf("invalid stocks qty: %s", cellValue(row, colStocksQty))
		}
		stocksQty = parsed
	}

	status := existing.Status
	if cellProvided(row, colStatus) {
		parsed := enums.ParseProductStatusToEnum(cellValue(row, colStatus))
		if parsed == enums.PRODUCT_STATUS_UNDEFINED {
			return UpdateProductInput{}, fmt.Errorf("invalid status: %s", cellValue(row, colStatus))
		}
		status = parsed.String()
	}

	externalLinks, err := mergeImportExternalLinks(row, headerMap, existing.ExternalLinks)
	if err != nil {
		return UpdateProductInput{}, err
	}

	return UpdateProductInput{
		BrandID:             brandID,
		Category:            mergeImportString(row, colCategory, existing.Category, importBlankBehavior(colCategory)),
		Subcategory:         mergeImportString(row, colSubcategory, existing.Subcategory, importBlankBehavior(colSubcategory)),
		Name:                mergeImportString(row, colProductName, existing.Name, importBlankBehavior(colProductName)),
		Description:         mergeImportString(row, colDescription, existing.Description, importBlankBehavior(colDescription)),
		Specs:               specs,
		Status:              status,
		UnitPriceWithoutVat: unitPriceWithoutVat,
		UnitPriceWithVat:    unitPriceWithVat,
		SalePriceWithoutVat: salePriceWithoutVat,
		SalePriceWithVat:    salePriceWithVat,
		SaleStartDate:       saleStartDate,
		SaleEndDate:         saleEndDate,
		StocksIn:            stocksIn,
		Stocks:              stocksQty,
		ExternalLinks:       externalLinks,
	}, nil
}

func (s *ProductBulkImportService) resolveBrandID(ctx context.Context, brandName string) (string, error) {
	brandID, err := s.product.ResolveBrandIDByName(ctx, brandName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("brand not found: %s", brandName)
		}
		return "", err
	}
	return brandID, nil
}

type rowPrices struct {
	unitPriceWithoutVat int64
	unitPriceWithVat    int64
	salePriceWithoutVat int64
	salePriceWithVat    int64
}

func parseRowPrices(row map[string]string) (rowPrices, error) {
	unitPriceWithVat, err := parseExportPrice(cellValue(row, colUnitPriceWithVat))
	if err != nil {
		return rowPrices{}, err
	}

	vatPercentage, err := strconv.ParseFloat(conf.Conf().Settings.VATPercentage, 64)
	if err != nil {
		vatPercentage = 0
	}
	unitPriceWithoutVat := int64(math.Round(float64(unitPriceWithVat) / (1 + vatPercentage/100)))

	prices := rowPrices{
		unitPriceWithoutVat: unitPriceWithoutVat,
		unitPriceWithVat:    unitPriceWithVat,
	}

	if raw := cellValue(row, colSalePriceWithVat); raw != "" {
		salePriceWithVat, parseErr := parseExportPrice(raw)
		if parseErr != nil {
			return rowPrices{}, parseErr
		}
		prices.salePriceWithVat = salePriceWithVat
		prices.salePriceWithoutVat = int64(math.Round(float64(salePriceWithVat) / (1 + vatPercentage/100)))
	}

	return prices, nil
}

func parseRowSpecs(row map[string]string) (ProductSpecsInput, error) {
	weightUnit, err := normalizeImportWeightUnit(cellValue(row, colWeightUnit))
	if err != nil {
		return ProductSpecsInput{}, err
	}

	return ProductSpecsInput{
		Colours:       cellValue(row, colColours),
		Sizes:         cellValue(row, colSizes),
		Segmentation:  cellValue(row, colSegmentation),
		PartNumber:    cellValue(row, colPartNumber),
		Power:         cellValue(row, colPower),
		Capacity:      cellValue(row, colCapacity),
		Weight:        cellValue(row, colWeight),
		WeightUnit:    weightUnit,
		ScopeOfSupply: cellValue(row, colScopeOfSupply),
	}, nil
}

func parseRowStocks(row map[string]string) (enums.StocksIn, int64, error) {
	stocksIn := enums.ParseStocksInToEnum(cellValue(row, colStocksIn))
	if stocksIn == enums.STOCKS_IN_UNDEFINED {
		return enums.STOCKS_IN_UNDEFINED, 0, fmt.Errorf("invalid stocks in: %s", cellValue(row, colStocksIn))
	}

	stocksQty, err := strconv.ParseInt(cellValue(row, colStocksQty), 10, 64)
	if err != nil {
		return enums.STOCKS_IN_UNDEFINED, 0, fmt.Errorf("invalid stocks qty: %s", cellValue(row, colStocksQty))
	}

	return stocksIn, stocksQty, nil
}

func parseRowSaleDates(row map[string]string, hasSale bool) (string, string, error) {
	if !hasSale {
		return "", "", nil
	}

	startDate, err := parseExportDate(cellValue(row, colSaleStartDate))
	if err != nil {
		return "", "", err
	}
	endDate, err := parseExportDate(cellValue(row, colSaleEndDate))
	if err != nil {
		return "", "", err
	}
	if startDate == "" || endDate == "" {
		return "", "", errors.New("sale start and end dates are required when sale price is set")
	}

	return startDate, endDate, nil
}

func validateRowExternalLinks(row map[string]string) ([]ExternalPlatformLinkInput, error) {
	links := externalLinksFromRow(row)
	if err := validateExternalPlatformLinks(links); err != nil {
		return nil, err
	}
	return links, nil
}

func parseProductImportFile(filename string, reader io.Reader) ([]string, [][]string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".csv":
		return parseProductImportCSV(reader)
	case ".xlsx":
		return parseProductImportXLSX(reader)
	default:
		return nil, nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func parseProductImportCSV(reader io.Reader) ([]string, [][]string, error) {
	r := csv.NewReader(reader)
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) < 2 {
		return nil, nil, errors.New("file has no data rows")
	}

	return records[0], records[1:], nil
}

func parseProductImportXLSX(reader io.Reader) ([]string, [][]string, error) {
	file, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, errors.New("xlsx has no sheets")
	}

	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, nil, err
	}
	if len(rows) < 2 {
		return nil, nil, errors.New("file has no data rows")
	}

	headers := rows[0]
	records := make([][]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		if isEmptyRow(row) {
			continue
		}
		values := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				values[i] = strings.TrimSpace(row[i])
			}
		}
		records = append(records, values)
	}

	return headers, records, nil
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func (s *ProductBulkImportService) ID() string {
	return "ProductBulkImport"
}

func (s *ProductBulkImportService) Log() {
	logs.Log().Info("[ProductBulkImportService] Loaded")
}

var _ IService = (*ProductBulkImportService)(nil)
