package utils

import (
	"fmt"
	"strings"

	"github.com/gosimple/slug"
)

func ProductSlug(
	brand string,
	category string,
	subcategory string,
	serial string,
	power string,
) string {
	brand = strings.ToLower(brand)

	if category == subcategory {
		subcategory = ""
	}

	serial = strings.ToLower(serial)
	serial = strings.TrimPrefix(serial, brand)
	serial = strings.TrimPrefix(serial, "-")
	serial = strings.Split(serial, "-")[0]

	power = strings.ReplaceAll(power, "-", "")
	power = strings.ReplaceAll(power, ",", "")
	power = strings.ReplaceAll(power, " ", "")

	s := strings.Join([]string{
		brand,
		category,
		subcategory,
		serial,
		power,
	}, "-")

	return slug.Make(s)
}

func BrandSlug(name string) string {
	return slug.Make(strings.TrimSpace(name))
}

func UniqueBrandSlug(base string, brandID int64, taken map[string]int64) string {
	if base == "" {
		base = fmt.Sprintf("brand-%d", brandID)
	}

	slugValue := base
	if existingID, ok := taken[slugValue]; ok && existingID != brandID {
		slugValue = fmt.Sprintf("%s-%d", base, brandID)
	}

	taken[slugValue] = brandID
	return slugValue
}
