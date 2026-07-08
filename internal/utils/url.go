package utils

import (
	"net/url"
	"strings"

	"cchoice/internal/constants"
	"cchoice/internal/errs"
)

func ValidateExternalURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return errs.ErrInvalidFormat
	}

	if !constants.ReExternalURL.MatchString(rawURL) {
		return errs.ErrInvalidFormat
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return errs.ErrInvalidFormat
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errs.ErrInvalidFormat
	}

	if parsed.Host == "" {
		return errs.ErrInvalidFormat
	}

	return nil
}
