package enums

import (
	"cchoice/internal/logs"

	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

//go:generate go tool stringer -type=Level -trimprefix=LEVEL_

type Level int

const (
	LEVEL_UNDEFINED Level = iota
	LEVEL_REGION
	LEVEL_PROVINCE
	LEVEL_MUNICIPALITY
	LEVEL_SUB_MUNICIPALITY
	LEVEL_CITY
	LEVEL_BARANGAY
)

func (l Level) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

func ParseLevelToEnum(e string) Level {
	switch e {
	case LEVEL_REGION.String():
		return LEVEL_REGION
	case LEVEL_PROVINCE.String():
		return LEVEL_PROVINCE
	case LEVEL_MUNICIPALITY.String():
		return LEVEL_MUNICIPALITY
	case LEVEL_SUB_MUNICIPALITY.String():
		return LEVEL_SUB_MUNICIPALITY
	case LEVEL_CITY.String():
		return LEVEL_CITY
	case LEVEL_BARANGAY.String():
		return LEVEL_BARANGAY
	default:
		logs.Log().Warn("Unhandled level", zap.Any("level", e))
		return LEVEL_UNDEFINED
	}
}

func ParseXLSXLevelToEnum(e string) Level {
	switch e {
	case "Reg":
		return LEVEL_REGION
	case "Prov":
		return LEVEL_PROVINCE
	case "Mun":
		return LEVEL_MUNICIPALITY
	case "SubMun":
		return LEVEL_SUB_MUNICIPALITY
	case "City":
		return LEVEL_CITY
	case "Bgy":
		return LEVEL_BARANGAY
	default:
		logs.Log().Warn("Unhandled level", zap.Any("level", e))
		return LEVEL_UNDEFINED
	}
}
