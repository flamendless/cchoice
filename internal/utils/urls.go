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

func MatchPath(path string, target string) bool {
	if conf.Conf().IsLocal() || conf.Conf().IsWeb() {
		return path == ("/cchoice" + target)
	}
	return path == target
}

func MetricsEvent(event string) string {
	return fmt.Sprintf("%s?event=%s", URL("/metrics/event"), event)
}
