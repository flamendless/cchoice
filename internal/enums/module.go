package enums

//go:generate go tool stringer -type=Module -trimprefix=MODULE_

type Module int

const (
	MODULE_UNDEFINED Module = iota
	MODULE_CATEGORY
	MODULE_PRODUCT
	MODULE_BRAND
)
