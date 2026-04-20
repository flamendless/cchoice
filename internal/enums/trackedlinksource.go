package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=TrackedLinkSource -trimprefix=TRACKED_LINK_SOURCE_

type TrackedLinkSource int

const (
	TRACKED_LINK_SOURCE_UNDEFINED TrackedLinkSource = iota
	TRACKED_LINK_SOURCE_QR
	TRACKED_LINK_SOURCE_EMAIL
	TRACKED_LINK_SOURCE_FACEBOOK
)

var AllTrackedLinkSources = []TrackedLinkSource{
	TRACKED_LINK_SOURCE_QR,
	TRACKED_LINK_SOURCE_EMAIL,
	TRACKED_LINK_SOURCE_FACEBOOK,
}

func ParseTrackedLinkSourceToEnum(s string) TrackedLinkSource {
	switch strings.ToUpper(s) {
	case TRACKED_LINK_SOURCE_QR.String():
		return TRACKED_LINK_SOURCE_QR
	case TRACKED_LINK_SOURCE_EMAIL.String():
		return TRACKED_LINK_SOURCE_EMAIL
	case TRACKED_LINK_SOURCE_FACEBOOK.String():
		return TRACKED_LINK_SOURCE_FACEBOOK
	default:
		return TRACKED_LINK_SOURCE_UNDEFINED
	}
}

func MustParseTrackedLinkSourceToEnum(s string) TrackedLinkSource {
	res := ParseTrackedLinkSourceToEnum(s)
	if res == TRACKED_LINK_SOURCE_UNDEFINED {
		panic(fmt.Sprintf("Unexpected TrackedLinkSource. Got '%s'", s))
	}
	return res
}
