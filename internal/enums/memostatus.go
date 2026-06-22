package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=MemoStatus -trimprefix=MEMO_STATUS_

type MemoStatus int

const (
	MEMO_STATUS_UNDEFINED MemoStatus = iota
	MEMO_STATUS_DRAFT
	MEMO_STATUS_PUBLISHED
	MEMO_STATUS_EXPIRED
	MEMO_STATUS_DELETED
)

var AllMemoStatuses = []MemoStatus{
	MEMO_STATUS_DRAFT,
	MEMO_STATUS_PUBLISHED,
	MEMO_STATUS_EXPIRED,
	MEMO_STATUS_DELETED,
}

func ParseMemoStatusToEnum(s string) MemoStatus {
	switch strings.ToUpper(s) {
	case MEMO_STATUS_DRAFT.String():
		return MEMO_STATUS_DRAFT
	case MEMO_STATUS_PUBLISHED.String():
		return MEMO_STATUS_PUBLISHED
	case MEMO_STATUS_EXPIRED.String():
		return MEMO_STATUS_EXPIRED
	case MEMO_STATUS_DELETED.String():
		return MEMO_STATUS_DELETED
	default:
		return MEMO_STATUS_UNDEFINED
	}
}

func MustParseMemoStatusToEnum(s string) MemoStatus {
	res := ParseMemoStatusToEnum(s)
	if res == MEMO_STATUS_UNDEFINED {
		panic(fmt.Sprintf("Unexpected MemoStatus. Got '%s'", s))
	}
	return res
}
