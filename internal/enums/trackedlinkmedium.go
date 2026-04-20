package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=TrackedLinkMedium -trimprefix=TRACKED_LINK_MEDIUM_

type TrackedLinkMedium int

const (
	TRACKED_LINK_MEDIUM_UNDEFINED TrackedLinkMedium = iota
	TRACKED_LINK_MEDIUM_SOCIAL
	TRACKED_LINK_MEDIUM_BUSINESS_CARD
)

var AllTrackedLinkMediums = []TrackedLinkMedium{
	TRACKED_LINK_MEDIUM_SOCIAL,
	TRACKED_LINK_MEDIUM_BUSINESS_CARD,
}

func ParseTrackedLinkMediumToEnum(s string) TrackedLinkMedium {
	switch strings.ToUpper(s) {
	case TRACKED_LINK_MEDIUM_SOCIAL.String():
		return TRACKED_LINK_MEDIUM_SOCIAL
	case TRACKED_LINK_MEDIUM_BUSINESS_CARD.String():
		return TRACKED_LINK_MEDIUM_BUSINESS_CARD
	default:
		return TRACKED_LINK_MEDIUM_UNDEFINED
	}
}

func MustParseTrackedLinkMediumToEnum(s string) TrackedLinkMedium {
	res := ParseTrackedLinkMediumToEnum(s)
	if res == TRACKED_LINK_MEDIUM_UNDEFINED {
		panic(fmt.Sprintf("Unexpected TrackedLinkMedium. Got '%s'", s))
	}
	return res
}
