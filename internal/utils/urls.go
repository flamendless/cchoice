package utils

import (
	"cchoice/internal/conf"
	"fmt"
)

func URL(path string) string {
	if conf.Conf().IsProd() {
		return path
	}
	return "/cchoice" + path
}

func URLWithSuccess(url string, message string) string {
	return URL(fmt.Sprintf("%s?success=%s", url, message))
}

func URLWithError(url string, message string) string {
	return URL(fmt.Sprintf("%s?error=%s", url, message))
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
