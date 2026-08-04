package datamigrate

import (
	"context"
	"fmt"
	"strings"

	"cchoice/internal/database"
)

type DoctorFinding struct {
	Severity string // "ok", "warn", "error"
	Title    string
	Detail   string
	Fix      string
}

func Doctor(ctx context.Context, db database.IService) ([]DoctorFinding, error) {
	findings := make([]DoctorFinding, 0)

	integrity, err := db.GetDB().QueryContext(ctx, "PRAGMA integrity_check")
	if err != nil {
		return nil, err
	}
	defer integrity.Close()

	integrityRows := make([]string, 0)
	for integrity.Next() {
		var row string
		if err := integrity.Scan(&row); err != nil {
			return nil, err
		}
		integrityRows = append(integrityRows, row)
	}
	if err := integrity.Err(); err != nil {
		return nil, err
	}

	if len(integrityRows) == 1 && integrityRows[0] == "ok" {
		findings = append(findings, DoctorFinding{
			Severity: "ok",
			Title:    "integrity_check",
			Detail:   "ok",
		})
	} else {
		findings = append(findings, DoctorFinding{
			Severity: "error",
			Title:    "integrity_check",
			Detail:   strings.Join(integrityRows, "; "),
			Fix:      "restore from backup (scripts/dbbackup.sh); do not run migrations on a corrupted file",
		})
	}

	pendingGoose, err := ListPendingGooseMigrations(ctx, db)
	if err != nil {
		return nil, err
	}
	if len(pendingGoose) == 0 {
		findings = append(findings, DoctorFinding{
			Severity: "ok",
			Title:    "goose migrations",
			Detail:   "all applied",
		})
	} else {
		var names []string
		for _, m := range pendingGoose {
			names = append(names, m.Filename)
		}
		findings = append(findings, DoctorFinding{
			Severity: "error",
			Title:    "goose migrations",
			Detail:   fmt.Sprintf("%d pending: %s", len(pendingGoose), strings.Join(names, ", ")),
			Fix:      "mage dbUp",
		})
	}

	state, err := GetState(ctx, db)
	if err != nil {
		return nil, err
	}
	switch {
	case state.PendingMigration:
		findings = append(findings, DoctorFinding{
			Severity: "error",
			Title:    "datamigrate table",
			Detail:   "tbl_datamigrate_applied missing",
			Fix:      "mage dbUp",
		})
	case len(state.PendingScripts) == 0:
		findings = append(findings, DoctorFinding{
			Severity: "ok",
			Title:    "post-migrate scripts",
			Detail:   "all applied",
		})
	default:
		var names []string
		for _, p := range state.PendingScripts {
			names = append(names, p.Name)
		}
		findings = append(findings, DoctorFinding{
			Severity: "error",
			Title:    "post-migrate scripts",
			Detail:   fmt.Sprintf("%d pending: %s", len(state.PendingScripts), strings.Join(names, ", ")),
			Fix:      "mage datamigrateup (or ./tmp/cchoiceprod datamigrate up)",
		})
	}

	findings = append(findings, checkInvoiceSchema(ctx, db))
	findings = append(findings, checkSlugBackfills(ctx, db)...)

	return findings, nil
}

func checkInvoiceSchema(ctx context.Context, db database.IService) DoctorFinding {
	rows, err := db.GetDB().QueryContext(ctx, "PRAGMA table_info(tbl_invoices)")
	if err != nil {
		return DoctorFinding{
			Severity: "warn",
			Title:    "invoice schema",
			Detail:   "tbl_invoices not readable: " + err.Error(),
			Fix:      "mage dbUp",
		}
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue any
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return DoctorFinding{
				Severity: "error",
				Title:    "invoice schema",
				Detail:   "failed reading table_info: " + err.Error(),
			}
		}
		cols[name] = true
	}
	if err := rows.Err(); err != nil {
		return DoctorFinding{
			Severity: "error",
			Title:    "invoice schema",
			Detail:   err.Error(),
		}
	}

	required := []string{"delivery_date", "payment_terms_value", "payment_terms_unit"}
	missing := make([]string, 0)
	for _, col := range required {
		if !cols[col] {
			missing = append(missing, col)
		}
	}
	if len(missing) == 0 {
		return DoctorFinding{
			Severity: "ok",
			Title:    "invoice schema",
			Detail:   "delivery/payment terms columns present",
		}
	}

	return DoctorFinding{
		Severity: "error",
		Title:    "invoice schema",
		Detail:   "missing columns: " + strings.Join(missing, ", "),
		Fix:      "mage dbUp (migration 20260803160000_invoice_dates_and_receipts.sql)",
	}
}

func checkSlugBackfills(ctx context.Context, db database.IService) []DoctorFinding {
	findings := make([]DoctorFinding, 0)

	productMissing, err := countQuery(ctx, db, `
		SELECT COUNT(*) FROM tbl_products
		WHERE deleted_at = '1970-01-01 00:00:00+00:00'
			AND (slug IS NULL OR slug = '')
	`)
	if err != nil {
		findings = append(findings, DoctorFinding{
			Severity: "warn",
			Title:    "product slugs",
			Detail:   "query failed: " + err.Error(),
		})
	} else {
		applied, err := scriptMarkedApplied(ctx, db, "populate_product_slugs")
		switch {
		case err != nil:
			findings = append(findings, DoctorFinding{
				Severity: "warn",
				Title:    "product slugs",
				Detail:   "could not read datamigrate applied: " + err.Error(),
			})
		case productMissing == 0:
			findings = append(findings, DoctorFinding{
				Severity: "ok",
				Title:    "product slugs",
				Detail:   "all products have slugs",
			})
		case applied:
			findings = append(findings, DoctorFinding{
				Severity: "error",
				Title:    "product slugs",
				Detail:   fmt.Sprintf("%d products missing slugs but populate_product_slugs is marked applied", productMissing),
				Fix:      "DELETE FROM tbl_datamigrate_applied WHERE name='populate_product_slugs'; then run populate_product_slugs --dry-run=false",
			})
		default:
			findings = append(findings, DoctorFinding{
				Severity: "warn",
				Title:    "product slugs",
				Detail:   fmt.Sprintf("%d products missing slugs", productMissing),
				Fix:      "populate_product_slugs --dry-run=false",
			})
		}
	}

	brandMissing, err := countQuery(ctx, db, `
		SELECT COUNT(*) FROM tbl_brands
		WHERE deleted_at = '1970-01-01 00:00:00+00:00'
			AND (slug IS NULL OR slug = '')
	`)
	if err != nil {
		findings = append(findings, DoctorFinding{
			Severity: "warn",
			Title:    "brand slugs",
			Detail:   "query failed (slug column may be missing): " + err.Error(),
			Fix:      "mage dbUp",
		})
	} else {
		applied, err := scriptMarkedApplied(ctx, db, "populate_brand_slugs")
		switch {
		case err != nil:
			findings = append(findings, DoctorFinding{
				Severity: "warn",
				Title:    "brand slugs",
				Detail:   "could not read datamigrate applied: " + err.Error(),
			})
		case brandMissing == 0:
			findings = append(findings, DoctorFinding{
				Severity: "ok",
				Title:    "brand slugs",
				Detail:   "all brands have slugs",
			})
		case applied:
			findings = append(findings, DoctorFinding{
				Severity: "error",
				Title:    "brand slugs",
				Detail:   fmt.Sprintf("%d brands missing slugs but populate_brand_slugs is marked applied", brandMissing),
				Fix:      "DELETE FROM tbl_datamigrate_applied WHERE name='populate_brand_slugs'; then run populate_brand_slugs --dry-run=false",
			})
		default:
			findings = append(findings, DoctorFinding{
				Severity: "warn",
				Title:    "brand slugs",
				Detail:   fmt.Sprintf("%d brands missing slugs", brandMissing),
				Fix:      "populate_brand_slugs --dry-run=false",
			})
		}
	}

	dupBrands, err := countQuery(ctx, db, `
		SELECT COUNT(*) FROM (
			SELECT slug FROM tbl_brands
			WHERE deleted_at = '1970-01-01 00:00:00+00:00'
				AND slug IS NOT NULL AND slug != ''
			GROUP BY slug HAVING COUNT(*) > 1
		)
	`)
	switch {
	case err != nil:
		findings = append(findings, DoctorFinding{
			Severity: "warn",
			Title:    "duplicate brand slugs",
			Detail:   "query failed: " + err.Error(),
		})
	case dupBrands == 0:
		findings = append(findings, DoctorFinding{
			Severity: "ok",
			Title:    "duplicate brand slugs",
			Detail:   "none",
		})
	default:
		findings = append(findings, DoctorFinding{
			Severity: "error",
			Title:    "duplicate brand slugs",
			Detail:   fmt.Sprintf("%d slug values shared by multiple brands", dupBrands),
			Fix:      "re-run populate_brand_slugs --dry-run=false after clearing bad slugs or unmarking the script",
		})
	}

	return findings
}

func countQuery(ctx context.Context, db database.IService, query string) (int64, error) {
	var n int64
	err := db.GetDB().QueryRowContext(ctx, query).Scan(&n)
	return n, err
}

func scriptMarkedApplied(ctx context.Context, db database.IService, name string) (bool, error) {
	exists, err := datamigrateTableExists(ctx, db)
	if err != nil || !exists {
		return false, err
	}
	var count int64
	err = db.GetDB().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tbl_datamigrate_applied WHERE name = ?`,
		name,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func FormatDoctor(findings []DoctorFinding) string {
	var b strings.Builder
	errors := 0
	warns := 0
	for _, f := range findings {
		switch f.Severity {
		case "error":
			errors++
		case "warn":
			warns++
		}
	}

	fmt.Fprintf(&b, "summary: %d error(s), %d warning(s)\n\n", errors, warns)
	for _, f := range findings {
		fmt.Fprintf(&b, "[%s] %s\n", f.Severity, f.Title)
		if f.Detail != "" {
			b.WriteString("  ")
			b.WriteString(f.Detail)
			b.WriteByte('\n')
		}
		if f.Fix != "" {
			b.WriteString("  fix: ")
			b.WriteString(f.Fix)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}
