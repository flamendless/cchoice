package services

import (
	"testing"

	"cchoice/internal/enums"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProductExportHeaderMap(t *testing.T) {
	t.Parallel()

	headers := []string{
		"Row Number",
		"Brand",
		"Serial Number",
		"Product Name",
	}
	_, err := parseProductExportHeaderMap(headers)
	assert.NoError(t, err)

	_, err = parseProductExportHeaderMap([]string{"brand", "product name"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serial number")
}

func TestExternalLinksFromRow(t *testing.T) {
	t.Parallel()

	row := map[string]string{
		colExternalLinkLazada: "https://lazada.test/item",
		colExternalLinkTiktok: "",
		colExternalLinkShopee: "https://shopee.test/item",
	}

	links := externalLinksFromRow(row)
	require.Len(t, links, 2)

	linkByPlatform := make(map[string]string, len(links))
	for _, link := range links {
		linkByPlatform[link.Platform] = link.URL
	}
	assert.Equal(t, "https://lazada.test/item", linkByPlatform[enums.EXTERNAL_PLATFORM_LAZADA.String()])
	assert.Equal(t, "https://shopee.test/item", linkByPlatform[enums.EXTERNAL_PLATFORM_SHOPEE.String()])
}

func TestParseExportPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{name: "plain number", input: "1234.56", want: 1235},
		{name: "with currency symbol", input: "₱1,234.56", want: 1235},
		{name: "empty", input: "", wantErr: true},
		{name: "zero", input: "0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseExportPrice(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseExportDate(t *testing.T) {
	t.Parallel()

	got, err := parseExportDate("2026-07-08 15:04:05")
	require.NoError(t, err)
	assert.Equal(t, "2026-07-08", got)

	got, err = parseExportDate("2026-07-08")
	require.NoError(t, err)
	assert.Equal(t, "2026-07-08", got)

	_, err = parseExportDate("not-a-date")
	assert.Error(t, err)
}

func TestNormalizeImportWeightUnit(t *testing.T) {
	t.Parallel()

	got, err := normalizeImportWeightUnit("KG")
	require.NoError(t, err)
	assert.Equal(t, "kg", got)

	got, err = normalizeImportWeightUnit("kg")
	require.NoError(t, err)
	assert.Equal(t, "kg", got)

	_, err = normalizeImportWeightUnit("invalid")
	assert.Error(t, err)
}

func TestProductExportRowToStrings_IncludesExternalLinks(t *testing.T) {
	t.Parallel()

	row := ProductExportRow{
		Brand:     "BOSCH",
		Serial:    "ABC123",
		Name:      "Test Product",
		LazadaURL: "https://lazada.test",
	}
	values := productExportRowToStrings(row, 1)
	assert.Equal(t, "https://lazada.test", values[len(productExportHeaders)-3])
}

func TestValidateCreateRowValues(t *testing.T) {
	t.Parallel()

	row := map[string]string{
		colBrand:            "BOSCH",
		colSerialNumber:     "SN-1",
		colProductName:      "Product",
		colDescription:      "Desc",
		colCategory:         "Tools",
		colSubcategory:      "Drills",
		colUnitPriceWithVat: "100",
		colColours:          "Red",
		colSizes:            "M",
		colSegmentation:     "Pro",
		colPartNumber:       "PN-1",
		colPower:            "18V",
		colCapacity:         "2Ah",
		colWeight:           "1.5",
		colWeightUnit:       "KG",
		colScopeOfSupply:    "Case",
		colStocksIn:         "WAREHOUSE",
		colStocksQty:        "10",
	}
	assert.NoError(t, validateCreateRowValues(row))

	delete(row, colDescription)
	assert.Error(t, validateCreateRowValues(row))
}
