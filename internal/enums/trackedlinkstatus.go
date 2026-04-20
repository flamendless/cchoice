package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=TrackedLinkStatus -trimprefix=TRACKED_LINK_STATUS_

type TrackedLinkStatus int

const (
	TRACKED_LINK_STATUS_UNDEFINED TrackedLinkStatus = iota
	TRACKED_LINK_STATUS_DRAFT
	TRACKED_LINK_STATUS_ACTIVE
	TRACKED_LINK_STATUS_DELETED
)

var AllTrackedLinkStatuses = []TrackedLinkStatus{
	TRACKED_LINK_STATUS_DRAFT,
	TRACKED_LINK_STATUS_ACTIVE,
	TRACKED_LINK_STATUS_DELETED,
}

func ParseTrackedLinkStatusToEnum(s string) TrackedLinkStatus {
	switch strings.ToUpper(s) {
	case TRACKED_LINK_STATUS_DRAFT.String():
		return TRACKED_LINK_STATUS_DRAFT
	case TRACKED_LINK_STATUS_ACTIVE.String():
		return TRACKED_LINK_STATUS_ACTIVE
	case TRACKED_LINK_STATUS_DELETED.String():
		return TRACKED_LINK_STATUS_DELETED
	default:
		return TRACKED_LINK_STATUS_UNDEFINED
	}
}

func MustParseTrackedLinkStatusToEnum(s string) TrackedLinkStatus {
	res := ParseTrackedLinkStatusToEnum(s)
	if res == TRACKED_LINK_STATUS_UNDEFINED {
		panic(fmt.Sprintf("Unexpected TrackedLinkStatus. Got '%s'", s))
	}
	return res
}
