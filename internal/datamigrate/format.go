package datamigrate

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"cchoice/internal/constants"
)

func FormatGooseVersion(version int64) string {
	if version < 10000000000000 {
		return strconv.FormatInt(version, 10)
	}

	s := padGooseVersionDigits(version)
	if len(s) < 14 {
		return strconv.FormatInt(version, 10)
	}

	t, err := time.Parse("20060102150405", s)
	if err != nil {
		return s
	}

	return t.Format(constants.DateTimeLayoutISO)
}

func padGooseVersionDigits(version int64) string {
	s := strconv.FormatInt(version, 10)
	if len(s) >= 14 {
		return s
	}
	return strings.Repeat("0", 14-len(s)) + s
}

func formatPendingScriptLine(order int, script PendingScript) string {
	return fmt.Sprintf(
		"%d. %d (%s) %s -> %s",
		order,
		script.AfterVersion,
		FormatGooseVersion(script.AfterVersion),
		script.Name,
		script.RunHint,
	)
}
