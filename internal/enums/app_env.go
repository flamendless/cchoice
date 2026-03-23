package enums

import "strings"

//go:generate go tool stringer -type=AppEnv -trimprefix=APP_ENV_

type AppEnv int

const (
	APP_ENV_UNDEFINED AppEnv = iota
	APP_ENV_LOCAL
	APP_ENV_WEB
	APP_ENV_PROD
)

func ParseAppEnvToEnum(e string) AppEnv {
	switch strings.ToUpper(e) {
	case APP_ENV_LOCAL.String():
		return APP_ENV_LOCAL
	case APP_ENV_WEB.String():
		return APP_ENV_WEB
	case APP_ENV_PROD.String():
		return APP_ENV_PROD
	default:
		return APP_ENV_UNDEFINED
	}
}
