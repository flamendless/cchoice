package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProductImportCSV(t *testing.T) {
	t.Parallel()

	csvData := `brand,serial number,product name
BOSCH,SN-1,Hammer
DELTA,SN-2,Drill
`
	headers, records, err := parseProductImportCSV(strings.NewReader(csvData))
	require.NoError(t, err)
	assert.Equal(t, []string{"brand", "serial number", "product name"}, headers)
	require.Len(t, records, 2)
	assert.Equal(t, "SN-1", records[0][1])
}

func TestParseProductImportCSV_NoDataRows(t *testing.T) {
	t.Parallel()

	_, _, err := parseProductImportCSV(strings.NewReader("brand,serial number,product name\n"))
	assert.Error(t, err)
}

func TestParseProductImportFile_UnsupportedExtension(t *testing.T) {
	t.Parallel()

	_, _, err := parseProductImportFile("products.txt", strings.NewReader("a,b\n1,2"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported file type")
}

func TestRowValuesToMap(t *testing.T) {
	t.Parallel()

	headers := []string{"Brand", "Serial Number"}
	values := []string{"BOSCH", "SN-1"}
	row := rowValuesToMap(headers, values)
	assert.Equal(t, "BOSCH", row["brand"])
	assert.Equal(t, "SN-1", row["serial number"])
}

func TestPickCell(t *testing.T) {
	t.Parallel()

	row := map[string]string{"brand": "NEW"}
	assert.Equal(t, "NEW", pickCell(row, "brand", "OLD"))
	assert.Equal(t, "OLD", pickCell(row, "status", "OLD"))
}

func TestIsEmptyRow(t *testing.T) {
	t.Parallel()

	assert.True(t, isEmptyRow([]string{"", "  "}))
	assert.False(t, isEmptyRow([]string{"", "x"}))
}

func TestParseRowSpecs(t *testing.T) {
	t.Parallel()

	row := map[string]string{
		colColours:       "Red",
		colSizes:         "L",
		colSegmentation:  "Pro",
		colPartNumber:    "PN",
		colPower:         "18V",
		colCapacity:      "2Ah",
		colWeight:        "1.2",
		colWeightUnit:    "KG",
		colScopeOfSupply: "Case",
	}
	specs, err := parseRowSpecs(row)
	require.NoError(t, err)
	assert.Equal(t, "Red", specs.Colours)
	assert.Equal(t, "kg", specs.WeightUnit)
	assert.Equal(t, "Case", specs.ScopeOfSupply)
}
