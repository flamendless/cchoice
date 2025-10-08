package enums

//go:generate go tool stringer -type=WeightUnit -trimprefix=WEIGHT_UNIT_

type WeightUnit int

const (
	WEIGHT_UNIT_UNDEFINED WeightUnit = iota
	WEIGHT_UNIT_KG
	WEIGHT_UNIT_G
	WEIGHT_UNIT_LB
	WEIGHT_UNIT_OZ
)

func ParseWeightUnitToEnum(unit string) WeightUnit {
	switch unit {
	case "kg", "KG", "kilogram", "kilograms":
		return WEIGHT_UNIT_KG
	case "g", "G", "gram", "grams":
		return WEIGHT_UNIT_G
	case "lb", "LB", "pound", "pounds", "lbs":
		return WEIGHT_UNIT_LB
	case "oz", "OZ", "ounce", "ounces":
		return WEIGHT_UNIT_OZ
	default:
		return WEIGHT_UNIT_UNDEFINED
	}
}

func (w WeightUnit) ToString() string {
	switch w {
	case WEIGHT_UNIT_KG:
		return "kg"
	case WEIGHT_UNIT_G:
		return "g"
	case WEIGHT_UNIT_LB:
		return "lb"
	case WEIGHT_UNIT_OZ:
		return "oz"
	default:
		return ""
	}
}
