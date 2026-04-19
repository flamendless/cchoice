package utils

import (
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
