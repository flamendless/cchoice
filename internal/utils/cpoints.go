package utils

import (
	"cchoice/internal/constants"
	"strconv"
)

func CalculateOrderEarnedCPoints(totalAmount int64) int64 {
	if totalAmount <= 0 {
		return 0
	}
	return (totalAmount*constants.CPointOrderRewardRatePercent + 5000) / 10000
}

func FormatEarnedCPoints(value int64) string {
	if value <= 0 {
		return "-"
	}
	return strconv.FormatInt(value, 10)
}
