package datamigrate_test

import (
	"context"
	"testing"

	"cchoice/internal/datamigrate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoctor_AllOK(t *testing.T) {
	t.Chdir("../../")

	db := newStubDB(t, true)
	ctx := context.Background()

	_, err := db.GetDB().Exec(`
		CREATE TABLE tbl_products (
			id INTEGER PRIMARY KEY,
			slug TEXT,
			deleted_at TEXT NOT NULL DEFAULT '1970-01-01 00:00:00+00:00'
		);
		CREATE TABLE tbl_brands (
			id INTEGER PRIMARY KEY,
			slug TEXT,
			deleted_at TEXT NOT NULL DEFAULT '1970-01-01 00:00:00+00:00'
		);
		CREATE TABLE tbl_invoices (id INTEGER PRIMARY KEY);
	`)
	require.NoError(t, err)

	findings, err := datamigrate.Doctor(ctx, db)
	require.NoError(t, err)
	require.NotEmpty(t, findings)

	for _, f := range findings {
		if f.Title == "integrity_check" {
			assert.Equal(t, "ok", f.Severity)
		}
	}
}

func TestDoctor_ProductSlugsMarkedButMissing(t *testing.T) {
	t.Chdir("../../")

	db := newStubDB(t, true)
	ctx := context.Background()

	_, err := db.GetDB().Exec(`
		CREATE TABLE tbl_products (
			id INTEGER PRIMARY KEY,
			slug TEXT,
			deleted_at TEXT NOT NULL DEFAULT '1970-01-01 00:00:00+00:00'
		);
		INSERT INTO tbl_products (id, slug) VALUES (1, '');
		INSERT INTO tbl_datamigrate_applied (name) VALUES ('populate_product_slugs');
	`)
	require.NoError(t, err)

	findings, err := datamigrate.Doctor(ctx, db)
	require.NoError(t, err)

	var productFinding datamigrate.DoctorFinding
	for _, f := range findings {
		if f.Title == "product slugs" {
			productFinding = f
			break
		}
	}
	assert.Equal(t, "error", productFinding.Severity)
	assert.Contains(t, productFinding.Detail, "missing slugs")
}
