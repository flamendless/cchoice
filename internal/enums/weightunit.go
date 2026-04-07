package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=WeightUnit -trimprefix=WEIGHT_UNIT_

type WeightUnit int

const (
	WEIGHT_UNIT_UNDEFINED WeightUnit = iota
	WEIGHT_UNIT_KG
	WEIGHT_UNIT_G
	WEIGHT_UNIT_LB
	WEIGHT_UNIT_OZ
)

var AllWeightUnits = []WeightUnit{
	WEIGHT_UNIT_KG,
	WEIGHT_UNIT_G,
	WEIGHT_UNIT_LB,
	WEIGHT_UNIT_OZ,
}

func ParseWeightUnitToEnum(unit string) WeightUnit {
	switch strings.ToUpper(unit) {
	case "KG", "KILOGRAM", "KILOGRAMS":
		return WEIGHT_UNIT_KG
	case "G", "GRAM", "GRAMS":
		return WEIGHT_UNIT_G
	case "LB", "LBS", "POUND", "POUNDS":
		return WEIGHT_UNIT_LB
	case "OZ", "OUNCE", "OUNCES":
		return WEIGHT_UNIT_OZ
	default:
		return WEIGHT_UNIT_UNDEFINED
	}
}

func (e WeightUnit) ToDB() string {
	switch e {
	case WEIGHT_UNIT_G:
		return "g"
	case WEIGHT_UNIT_KG:
		return "kg"
	case WEIGHT_UNIT_LB:
		return "lb"
	case WEIGHT_UNIT_OZ:
		return "oz"
	default:
		panic(fmt.Sprintf("unexpected enums.WeightUnit: %#v", e))
	}
}
