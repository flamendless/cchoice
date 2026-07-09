package utils

import (
	"cchoice/internal/conf"
	"fmt"
	"net/url"
	"strings"
)

func FullURL(path string) string {
	return siteBaseURL() + URL(path)
}

func SiteURL(path string) string {
	return siteBaseURL() + URL(path)
}

func siteBaseURL() string {
	base := strings.TrimSuffix(conf.Conf().Server.Address, "/")
	if strings.HasPrefix(base, "http://") || strings.HasPrefix(base, "https://") {
		return base
	}
	scheme := "https://"
	if conf.Conf().IsLocal() {
		scheme = "http://"
	}
	return scheme + base
}

func URL(path string) string {
	if conf.Conf().IsProd() {
		return path
	}
	return "/cchoice" + path
}

func URLf(path string, args ...any) string {
	return URL(fmt.Sprintf(path, args...))
}

func URLWithParams(path string, params map[string]string) string {
	return URL(appendQueryParams(path, params))
}

func appendQueryParams(path string, params map[string]string) string {
	base := path
	hasQuery := strings.Contains(path, "?")
	for k, v := range params {
		sep := "?"
		if hasQuery {
			sep = "&"
		}
		base = fmt.Sprintf("%s%s%s=%s", base, sep, url.QueryEscape(k), url.QueryEscape(v))
		hasQuery = true
	}
	return base
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
	return fmt.Sprintf("%s?event=%s", URL("/collect/event"), event)
}
