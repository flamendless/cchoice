package datamigrate

import (
	"os"

	"cchoice/internal/storage"
)

type Script struct {
	Name         string
	AfterVersion int64
	Command      string
	RequiredArgs []string
	RunHint      string
	When         func() bool
}

var Scripts = []Script{
	{
		Name:         "apply_discount:sale_2025",
		AfterVersion: 20260109164336,
		Command:      "apply_discount",
		RequiredArgs: []string{"scripts/csv/sale_2025.csv"},
		RunHint:      "./tmp/main apply_discount -i scripts/csv/sale_2025.csv --dry-run=false",
	},
	{
		Name:         "populate_product_images_cdn",
		AfterVersion: 20260327151912,
		Command:      "populate_product_images_cdn",
		RunHint:      "./tmp/main populate_product_images_cdn --dry-run=false",
		When: func() bool {
			provider := os.Getenv("STORAGE_PROVIDER")
			if provider == "" {
				provider = "LOCAL"
			}
			return provider == storage.STORAGE_PROVIDER_CLOUDFLARE_IMAGES.String()
		},
	},
	{
		Name:         "populate_product_slugs",
		AfterVersion: 20260418065629,
		Command:      "populate_product_slugs",
		RunHint:      "./tmp/main populate_product_slugs --dry-run=false",
	},
	{
		Name:         "populate_brand_slugs",
		AfterVersion: 20260710120000,
		Command:      "populate_brand_slugs",
		RunHint:      "./tmp/main populate_brand_slugs --dry-run=false",
	},
}
