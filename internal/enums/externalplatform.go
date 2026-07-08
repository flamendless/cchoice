package enums

import (
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=ExternalPlatform -trimprefix=EXTERNAL_PLATFORM_

type ExternalPlatform int

const (
	EXTERNAL_PLATFORM_UNDEFINED ExternalPlatform = iota
	EXTERNAL_PLATFORM_LAZADA
	EXTERNAL_PLATFORM_TIKTOK
	EXTERNAL_PLATFORM_SHOPEE
)

var AllExternalPlatforms = []ExternalPlatform{
	EXTERNAL_PLATFORM_LAZADA,
	EXTERNAL_PLATFORM_TIKTOK,
	EXTERNAL_PLATFORM_SHOPEE,
}

func ParseExternalPlatformToEnum(s string) ExternalPlatform {
	switch strings.ToUpper(s) {
	case EXTERNAL_PLATFORM_LAZADA.String():
		return EXTERNAL_PLATFORM_LAZADA
	case EXTERNAL_PLATFORM_TIKTOK.String():
		return EXTERNAL_PLATFORM_TIKTOK
	case EXTERNAL_PLATFORM_SHOPEE.String():
		return EXTERNAL_PLATFORM_SHOPEE
	default:
		return EXTERNAL_PLATFORM_UNDEFINED
	}
}

func MustParseExternalPlatformToEnum(s string) ExternalPlatform {
	res := ParseExternalPlatformToEnum(s)
	if res == EXTERNAL_PLATFORM_UNDEFINED {
		panic(fmt.Sprintf("Unexpected ExternalPlatform. Got '%s'", s))
	}
	return res
}
