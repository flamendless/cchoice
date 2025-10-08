package utils

import (
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
	"database/sql"
	"fmt"
	"strconv"
)

func ConvertWeightToKg(weight sql.NullFloat64, unit sql.NullString) (float64, error) {
	if !weight.Valid {
		return 1.0, nil
	}

	weightValue := weight.Float64
	if weightValue <= 0 {
		return 1.0, nil
	}

	if !unit.Valid || unit.String == "" {
		return weightValue, nil
	}

	weightUnit := enums.ParseWeightUnitToEnum(unit.String)
	switch weightUnit {
	case enums.WEIGHT_UNIT_KG:
		return weightValue, nil
	case enums.WEIGHT_UNIT_G:
		return weightValue / 1000.0, nil
	case enums.WEIGHT_UNIT_LB:
		return weightValue * 0.453592, nil
	case enums.WEIGHT_UNIT_OZ:
		return weightValue * 0.0283495, nil
	default:
		return weightValue, nil
	}
}

func CalculateTotalWeightFromCheckoutLines(checkoutLines []queries.GetCheckoutLinesByCheckoutIDRow) (string, error) {
	totalWeightKg := 0.0

	for _, checkoutLine := range checkoutLines {
		itemWeightKg, err := ConvertWeightToKg(checkoutLine.Weight, checkoutLine.WeightUnit)
		if err != nil {
			return "", fmt.Errorf("failed to convert weight: %w", err)
		}

		totalWeightKg += itemWeightKg * float64(checkoutLine.Quantity)
	}

	return strconv.FormatFloat(totalWeightKg, 'f', 2, 64), nil
}
