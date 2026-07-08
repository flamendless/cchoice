package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/logs"

	"go.uber.org/zap"
)

var productImportDiffFields = []string{
	colBrand,
	colProductName,
	colStatus,
	colCategory,
	colSubcategory,
	colDescription,
	colUnitPriceWithVat,
	colSalePriceWithVat,
	colSaleStartDate,
	colSaleEndDate,
	colColours,
	colSizes,
	colSegmentation,
	colPartNumber,
	colPower,
	colCapacity,
	colWeight,
	colWeightUnit,
	colScopeOfSupply,
	colStocksIn,
	colStocksQty,
	colExternalLinkLazada,
	colExternalLinkTiktok,
	colExternalLinkShopee,
}

func (s *ProductBulkImportService) PreviewFromReader(
	ctx context.Context,
	filename string,
	reader io.Reader,
) (*BulkImportPreview, *ProductImportSessionData, error) {
	headers, records, err := parseProductImportFile(filename, reader)
	if err != nil {
		return nil, nil, err
	}

	if _, err := parseProductExportHeaderMap(headers); err != nil {
		return nil, nil, err
	}
	headerMap, _ := parseProductExportHeaderMap(headers)

	preview := &BulkImportPreview{
		Rows:      make([]BulkImportPreviewRow, 0, len(records)),
		TotalRows: len(records),
	}

	for i, values := range records {
		line := i + 2
		row := rowValuesToMap(headers, values)
		previewRow, rowErr := s.previewRow(ctx, line, row, headerMap)
		if rowErr != nil {
			return nil, nil, rowErr
		}
		preview.Rows = append(preview.Rows, previewRow)

		switch previewRow.Action {
		case "create":
			preview.CreateCount++
		case "update":
			preview.UpdateCount++
		case "unchanged":
			preview.UnchangedCount++
		case "error":
			preview.ErrorCount++
		}
		if previewRow.Selectable {
			preview.SelectableCount++
		}
	}

	sortBulkImportPreviewRows(preview.Rows)

	return preview, &ProductImportSessionData{
		Headers: headers,
		Records: records,
	}, nil
}

func (s *ProductBulkImportService) previewRow(
	ctx context.Context,
	line int,
	row map[string]string,
	headerMap map[string]int,
) (BulkImportPreviewRow, error) {
	serial := cellValue(row, colSerialNumber)
	productName := cellValue(row, colProductName)

	if serial == "" {
		return BulkImportPreviewRow{
			Line:        line,
			Serial:      serial,
			ProductName: productName,
			Action:      "error",
			Error:       "serial number is required",
		}, nil
	}

	existing, err := s.product.GetBySerial(ctx, serial)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return BulkImportPreviewRow{}, err
	}

	if errors.Is(err, sql.ErrNoRows) {
		if err := validateCreateRowValues(row); err != nil {
			return BulkImportPreviewRow{
				Line:        line,
				Serial:      serial,
				ProductName: productName,
				Action:      "error",
				Error:       err.Error(),
			}, nil
		}

		after, snapErr := s.snapshotAfterCreate(ctx, row)
		if snapErr != nil {
			return BulkImportPreviewRow{
				Line:        line,
				Serial:      serial,
				ProductName: productName,
				Action:      "error",
				Error:       snapErr.Error(),
			}, nil
		}

		return BulkImportPreviewRow{
			Line:        line,
			Serial:      serial,
			ProductName: productName,
			Action:      "create",
			Changes:     diffSnapshots(nil, after),
			Selectable:  true,
		}, nil
	}

	productID := s.product.EncodeID(existing.ID)
	existingEdit, err := s.product.GetByIDForEdit(ctx, productID)
	if err != nil {
		return BulkImportPreviewRow{}, err
	}

	before := snapshotFromForEdit(existingEdit)
	after, err := s.snapshotAfterUpdate(ctx, row, existingEdit, headerMap)
	if err != nil {
		return BulkImportPreviewRow{
			Line:        line,
			Serial:      serial,
			ProductName: pickCell(row, colProductName, existingEdit.Name),
			Action:      "error",
			Error:       err.Error(),
		}, nil
	}

	changes := diffSnapshots(before, after)
	if len(changes) == 0 {
		return BulkImportPreviewRow{
			Line:        line,
			Serial:      serial,
			ProductName: existingEdit.Name,
			Action:      "unchanged",
		}, nil
	}

	return BulkImportPreviewRow{
		Line:        line,
		Serial:      serial,
		ProductName: pickCell(row, colProductName, existingEdit.Name),
		Action:      "update",
		Changes:     changes,
		Selectable:  true,
	}, nil
}

func (s *ProductBulkImportService) snapshotAfterCreate(
	ctx context.Context,
	row map[string]string,
) (map[string]string, error) {
	input, err := s.buildCreateInput(ctx, row)
	if err != nil {
		return nil, err
	}

	return snapshotFromCreateInput(input, cellValue(row, colBrand)), nil
}

func (s *ProductBulkImportService) snapshotAfterUpdate(
	ctx context.Context,
	row map[string]string,
	existing *ProductForEdit,
	headerMap map[string]int,
) (map[string]string, error) {
	input, err := s.buildUpdateInput(ctx, row, existing, headerMap)
	if err != nil {
		return nil, err
	}

	brandName := mergeImportString(row, colBrand, existing.BrandName, importBlankBehavior(colBrand))
	return snapshotFromUpdateInput(input, brandName), nil
}

func snapshotFromForEdit(p *ProductForEdit) map[string]string {
	links := externalLinksToColumnMap(p.ExternalLinks)
	return map[string]string{
		colBrand:              p.BrandName,
		colProductName:        p.Name,
		colStatus:             p.Status,
		colCategory:           p.Category,
		colSubcategory:        p.Subcategory,
		colDescription:        p.Description,
		colUnitPriceWithVat:   formatImportPricePesos(p.UnitPriceWithVat / 100),
		colSalePriceWithVat:   formatImportPricePesos(p.SalePriceWithVat / 100),
		colSaleStartDate:      p.SaleStartDate,
		colSaleEndDate:        p.SaleEndDate,
		colColours:            p.Specs.Colours,
		colSizes:              p.Specs.Sizes,
		colSegmentation:       p.Specs.Segmentation,
		colPartNumber:         p.Specs.PartNumber,
		colPower:              p.Specs.Power,
		colCapacity:           p.Specs.Capacity,
		colWeight:             p.Specs.Weight,
		colWeightUnit:         p.Specs.WeightUnit,
		colScopeOfSupply:      p.Specs.ScopeOfSupply,
		colStocksIn:           p.StocksIn.String(),
		colStocksQty:          strconv.FormatInt(p.Stocks, 10),
		colExternalLinkLazada: links[colExternalLinkLazada],
		colExternalLinkTiktok: links[colExternalLinkTiktok],
		colExternalLinkShopee: links[colExternalLinkShopee],
	}
}

func snapshotFromCreateInput(input CreateProductInput, brandName string) map[string]string {
	links := externalLinksToColumnMap(input.ExternalLinks)
	return map[string]string{
		colBrand:              brandName,
		colProductName:        input.Name,
		colStatus:             enums.PRODUCT_STATUS_DRAFT.String(),
		colCategory:           input.Category,
		colSubcategory:        input.Subcategory,
		colDescription:        input.Description,
		colUnitPriceWithVat:   formatImportPricePesos(input.UnitPriceWithVat),
		colSalePriceWithVat:   formatImportPricePesos(input.SalePriceWithVat),
		colSaleStartDate:      input.SaleStartDate,
		colSaleEndDate:        input.SaleEndDate,
		colColours:            input.Specs.Colours,
		colSizes:              input.Specs.Sizes,
		colSegmentation:       input.Specs.Segmentation,
		colPartNumber:         input.Specs.PartNumber,
		colPower:              input.Specs.Power,
		colCapacity:           input.Specs.Capacity,
		colWeight:             input.Specs.Weight,
		colWeightUnit:         input.Specs.WeightUnit,
		colScopeOfSupply:      input.Specs.ScopeOfSupply,
		colStocksIn:           input.StocksIn.String(),
		colStocksQty:          strconv.FormatInt(input.Stocks, 10),
		colExternalLinkLazada: links[colExternalLinkLazada],
		colExternalLinkTiktok: links[colExternalLinkTiktok],
		colExternalLinkShopee: links[colExternalLinkShopee],
	}
}

func snapshotFromUpdateInput(input UpdateProductInput, brandName string) map[string]string {
	links := externalLinksToColumnMap(input.ExternalLinks)
	return map[string]string{
		colBrand:              brandName,
		colProductName:        input.Name,
		colStatus:             input.Status,
		colCategory:           input.Category,
		colSubcategory:        input.Subcategory,
		colDescription:        input.Description,
		colUnitPriceWithVat:   formatImportPricePesos(input.UnitPriceWithVat),
		colSalePriceWithVat:   formatImportPricePesos(input.SalePriceWithVat),
		colSaleStartDate:      input.SaleStartDate,
		colSaleEndDate:        input.SaleEndDate,
		colColours:            input.Specs.Colours,
		colSizes:              input.Specs.Sizes,
		colSegmentation:       input.Specs.Segmentation,
		colPartNumber:         input.Specs.PartNumber,
		colPower:              input.Specs.Power,
		colCapacity:           input.Specs.Capacity,
		colWeight:             input.Specs.Weight,
		colWeightUnit:         input.Specs.WeightUnit,
		colScopeOfSupply:      input.Specs.ScopeOfSupply,
		colStocksIn:           input.StocksIn.String(),
		colStocksQty:          strconv.FormatInt(input.Stocks, 10),
		colExternalLinkLazada: links[colExternalLinkLazada],
		colExternalLinkTiktok: links[colExternalLinkTiktok],
		colExternalLinkShopee: links[colExternalLinkShopee],
	}
}

func externalLinksToColumnMap(links []ExternalPlatformLinkInput) map[string]string {
	result := map[string]string{
		colExternalLinkLazada: "",
		colExternalLinkTiktok: "",
		colExternalLinkShopee: "",
	}
	for _, link := range links {
		switch strings.ToUpper(link.Platform) {
		case enums.EXTERNAL_PLATFORM_LAZADA.String():
			result[colExternalLinkLazada] = link.URL
		case enums.EXTERNAL_PLATFORM_TIKTOK.String():
			result[colExternalLinkTiktok] = link.URL
		case enums.EXTERNAL_PLATFORM_SHOPEE.String():
			result[colExternalLinkShopee] = link.URL
		}
	}
	return result
}

func diffSnapshots(before, after map[string]string) []BulkImportFieldChange {
	changes := make([]BulkImportFieldChange, 0)
	for _, field := range productImportDiffFields {
		beforeVal := ""
		if before != nil {
			beforeVal = before[field]
		}
		afterVal := after[field]
		if beforeVal == afterVal {
			continue
		}
		changes = append(changes, BulkImportFieldChange{
			Field:  field,
			Before: displayImportValue(beforeVal),
			After:  displayImportValue(afterVal),
		})
	}
	return changes
}

func sortBulkImportPreviewRows(rows []BulkImportPreviewRow) {
	slices.SortFunc(rows, func(a, b BulkImportPreviewRow) int {
		if o := bulkImportPreviewRowPriority(a) - bulkImportPreviewRowPriority(b); o != 0 {
			return o
		}
		return a.Line - b.Line
	})
}

func bulkImportPreviewRowPriority(row BulkImportPreviewRow) int {
	switch row.Action {
	case "error":
		return 0
	case "create":
		return 1
	case "update":
		return 2
	case "unchanged":
		return 3
	default:
		return 4
	}
}

func displayImportValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

func formatImportPricePesos(pesos int64) string {
	if pesos == 0 {
		return ""
	}
	return strconv.FormatInt(pesos, 10)
}

func (s *ProductBulkImportService) ApplySelected(
	ctx context.Context,
	staffID string,
	data *ProductImportSessionData,
	selectedLines []int,
) (*BulkImportResult, error) {
	if data == nil || len(data.Headers) == 0 {
		return nil, errors.New("import preview expired, please upload the file again")
	}

	if _, err := parseProductExportHeaderMap(data.Headers); err != nil {
		return nil, err
	}
	headerMap, _ := parseProductExportHeaderMap(data.Headers)

	selected := make(map[int]struct{}, len(selectedLines))
	for _, line := range selectedLines {
		selected[line] = struct{}{}
	}

	result := &BulkImportResult{}
	logResult := "success"
	defer func() {
		if err := s.staffLog.CreateLog(
			ctx,
			staffID,
			constants.ActionUpdate,
			constants.ModuleProductsBulkImport,
			logResult,
			nil,
		); err != nil {
			logs.LogCtx(ctx).Error("[ProductBulkImportService] failed to log bulk import", zap.Error(err))
		}
	}()

	for i, values := range data.Records {
		line := i + 2
		if _, ok := selected[line]; !ok {
			continue
		}

		row := rowValuesToMap(data.Headers, values)
		serial := cellValue(row, colSerialNumber)
		if serial == "" {
			result.Failed++
			result.Errors = append(result.Errors, BulkImportRowError{
				Line:   line,
				Serial: serial,
				Reason: "serial number is required",
			})
			continue
		}

		action, rowErr := s.upsertRow(ctx, staffID, row, headerMap)
		if rowErr != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkImportRowError{
				Line:   line,
				Serial: serial,
				Reason: rowErr.Error(),
			})
			continue
		}

		switch action {
		case "created":
			result.Created++
		case "updated":
			result.Updated++
		}
	}

	logResult = fmt.Sprintf(
		"success. created %d, updated %d, failed %d",
		result.Created,
		result.Updated,
		result.Failed,
	)
	return result, nil
}
