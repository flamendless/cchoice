package utils

import (
	"cchoice/internal/conf"
	"fmt"
	"net/url"
)

func URL(path string) string {
	if conf.Conf().IsProd() {
		return path
	}
	return "/cchoice" + path
}

func URLf(path string, args ...any) string {
	return URL(fmt.Sprintf(path, args...))
}

func URLWithSuccess(path string, message string) string {
	message = url.QueryEscape(message)
	return URL(fmt.Sprintf("%s?success=%s", path, message))
}

func URLWithError(path string, message string) string {
	message = url.QueryEscape(message)
	return URL(fmt.Sprintf("%s?error=%s", path, message))
}

func MatchPath(path string, target string) bool {
	if conf.Conf().IsLocal() || conf.Conf().IsWeb() {
		return path == ("/cchoice" + target)
	}
	return path == target
}

func MetricsEvent(event string) string {
	return fmt.Sprintf("%s?event=%s", URL("/metrics/event"), event)
}
