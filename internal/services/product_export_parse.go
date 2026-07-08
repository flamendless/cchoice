package services

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"cchoice/internal/constants"
	"cchoice/internal/enums"
)

func parseProductExportHeaderMap(headers []string) (map[string]int, error) {
	col := make(map[string]int, len(headers))
	for i, h := range headers {
		key := strings.ToLower(strings.TrimSpace(h))
		if key == "" {
			continue
		}
		if _, exists := col[key]; exists {
			return nil, fmt.Errorf("duplicate column header: %s", h)
		}
		col[key] = i
	}

	for _, required := range productExportRequiredImportColumns {
		if _, ok := col[required]; !ok {
			return nil, fmt.Errorf("missing required column: %s", required)
		}
	}

	return col, nil
}

func rowValuesToMap(headers []string, values []string) map[string]string {
	result := make(map[string]string, len(headers))
	for i, h := range headers {
		key := strings.ToLower(strings.TrimSpace(h))
		if key == "" {
			continue
		}
		val := ""
		if i < len(values) {
			val = strings.TrimSpace(values[i])
		}
		result[key] = val
	}
	return result
}

func cellValue(row map[string]string, col string) string {
	return strings.TrimSpace(row[col])
}

func externalLinksFromRow(row map[string]string) []ExternalPlatformLinkInput {
	links := make([]ExternalPlatformLinkInput, 0, len(productExternalPlatformColumns))
	for _, item := range productExternalPlatformColumns {
		url := cellValue(row, item.Column)
		if url == "" {
			continue
		}
		links = append(links, ExternalPlatformLinkInput{
			Platform: item.Platform.String(),
			URL:      url,
		})
	}
	return links
}

func parseExportPrice(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("price is empty")
	}

	cleaned := constants.ReExportPriceCleanup.ReplaceAllString(s, "")
	if cleaned == "" || cleaned == "." || cleaned == "-" {
		return 0, fmt.Errorf("invalid price: %s", s)
	}

	price, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid price: %s", s)
	}
	if price <= 0 {
		return 0, fmt.Errorf("price must be positive: %s", s)
	}

	return int64(math.Round(price)), nil
}

func validateCreateRowValues(row map[string]string) error {
	for _, required := range productExportRequiredCreateColumns {
		if cellValue(row, required) == "" {
			return fmt.Errorf("missing required field: %s", required)
		}
	}
	return nil
}

func parseExportDate(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}

	layouts := []string{
		constants.DateTimeLayoutISO,
		constants.DateLayoutISO,
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format(constants.DateLayoutISO), nil
		}
	}
	return "", fmt.Errorf("invalid date: %s", s)
}

func normalizeImportWeightUnit(unit string) (string, error) {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return "", nil
	}
	parsed := enums.ParseWeightUnitToEnum(unit)
	if parsed == enums.WEIGHT_UNIT_UNDEFINED {
		return "", fmt.Errorf("invalid weight unit: %s", unit)
	}
	return parsed.ToDB(), nil
}
