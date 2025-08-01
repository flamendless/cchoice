package utils

import (
	"cchoice/internal/errs"
	"strconv"
)

func GetLimit(limit string) (int64, error) {
	if limit == "" {
		limit = "100"
	}
	res, err := strconv.Atoi(limit)
	if err != nil || res <= 0 {
		return 0, errs.ErrInvalidParamLimit
	}
	return int64(res), nil
}
