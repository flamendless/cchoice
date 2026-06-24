package services

import (
	"testing"

	"cchoice/internal/enums"

	"github.com/stretchr/testify/assert"
)

func TestSortProductExportRows_ByProductTitleAsc(t *testing.T) {
	t.Parallel()

	rows := []ProductExportRow{
		{Name: "Zebra Tool", Serial: "1"},
		{Name: "Alpha Drill", Serial: "2"},
		{Name: "Mango Saw", Serial: "3"},
	}

	sortProductExportRows(rows, enums.PRODUCT_EXPORT_SORT_COLUMN_PRODUCT_TITLE, enums.PRODUCT_EXPORT_SORT_DIRECTION_ASC)

	assert.Equal(t, "Alpha Drill", rows[0].Name)
	assert.Equal(t, "Mango Saw", rows[1].Name)
	assert.Equal(t, "Zebra Tool", rows[2].Name)
}

func TestSortProductExportRows_ByUpdatedAtDesc(t *testing.T) {
	t.Parallel()

	rows := []ProductExportRow{
		{Name: "old", UpdatedAt: "2024-01-01 00:00:00", Serial: "1"},
		{Name: "new", UpdatedAt: "2026-01-01 00:00:00", Serial: "2"},
		{Name: "mid", UpdatedAt: "2025-01-01 00:00:00", Serial: "3"},
	}

	sortProductExportRows(rows, enums.PRODUCT_EXPORT_SORT_COLUMN_UPDATED_AT, enums.PRODUCT_EXPORT_SORT_DIRECTION_DESC)

	assert.Equal(t, "new", rows[0].Name)
	assert.Equal(t, "mid", rows[1].Name)
	assert.Equal(t, "old", rows[2].Name)
}

func TestParseProductExportSortParams(t *testing.T) {
	t.Parallel()

	col := enums.ParseProductExportSortColumnToEnum("product_title")
	assert.Equal(t, enums.PRODUCT_EXPORT_SORT_COLUMN_PRODUCT_TITLE, col)

	dir := enums.ParseProductExportSortDirectionToEnum("asc")
	assert.Equal(t, enums.PRODUCT_EXPORT_SORT_DIRECTION_ASC, dir)
}
