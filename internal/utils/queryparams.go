package utils

import (
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

func GetListingSortFromQuery(q url.Values) (sortBy string, sortDir enums.ListingSortDirection) {
	sortBy = strings.ToUpper(q.Get("sort_by"))
	sortDir = enums.ParseListingSortDirection(q.Get("sort_dir"))
	return sortBy, sortDir
}

func ParseListingSortQuery(q url.Values, allowedSortBy ...string) (sortBy string, sortDir enums.ListingSortDirection, err error) {
	sortBy, sortDir = GetListingSortFromQuery(q)
	if sortBy != "" && !slices.Contains(allowedSortBy, sortBy) {
		return sortBy, enums.LISTING_SORT_DIRECTION_UNDEFINED, errs.ErrEnumInvalid
	}
	rawSortDir := strings.ToUpper(strings.TrimSpace(q.Get("sort_dir")))
	if rawSortDir != "" && sortDir == enums.LISTING_SORT_DIRECTION_UNDEFINED {
		return sortBy, sortDir, errs.ErrEnumInvalid
	}
	return sortBy, sortDir, nil
}

func NormalizeListingSort(sortBy string, sortDir enums.ListingSortDirection, defaultSortBy string) (string, enums.ListingSortDirection) {
	if sortBy == "" {
		sortBy = defaultSortBy
	}
	if sortDir == enums.LISTING_SORT_DIRECTION_UNDEFINED {
		sortDir = enums.LISTING_SORT_DIRECTION_DESC
	}
	return sortBy, sortDir
}

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
