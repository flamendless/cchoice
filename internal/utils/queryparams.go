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
	if err != nil {
		return 0, errs.ERR_INVALID_PARAMS
	}
	return int64(res), nil
}
