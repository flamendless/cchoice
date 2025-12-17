package changelogs

import (
	"bufio"
	"io"
	"strings"
)

func Parse(r io.Reader, appenv string, limit int) ([]ChangeLog, error) {
	scanner := bufio.NewScanner(r)

	var (
		count        int
		currentLog   *ChangeLog
		currentSect  *ChangeSection
	)
	logs := make([]ChangeLog, 0, limit*2)

	flushLog := func() {
		if currentLog == nil {
			return
		}

		sections := currentLog.Sections[:0]
		for _, s := range currentLog.Sections {
			if len(s.Items) > 0 {
				sections = append(sections, s)
			}
		}
		currentLog.Sections = sections

		if len(currentLog.Sections) > 0 {
			logs = append(logs, *currentLog)
		}
	}

	for scanner.Scan() {
		if count >= limit && currentSect != nil {
			break
		}

		line := strings.TrimSpace(scanner.Text())

		if line == "## [Unreleased]" {
			continue
		}

		// ## [release-v1.0.3] - 2025-12-17
		if strings.HasPrefix(line, "## [") {
			flushLog()

			version, date := parseHeader(line)
			switch appenv {
			case "local":
				if !strings.HasPrefix(version, "dev-v") {
					currentLog = nil
					continue
				}
			case "prod":
				if !strings.HasPrefix(version, "release-v") {
					currentLog = nil
					continue
				}
			}

			currentLog = &ChangeLog{
				Version: version,
				Date:    date,
				Anchor:  strings.ToLower(strings.ReplaceAll(version, ".", "-")),
			}
			currentSect = nil
			count++
			continue
		}

		// ### Server / ### Web / ### Feature / ### Docs
		if strings.HasPrefix(line, "### ") && currentLog != nil {
			section := ChangeSection{
				Title: strings.TrimPrefix(line, "### "),
			}
			currentLog.Sections = append(currentLog.Sections, section)
			currentSect = &currentLog.Sections[len(currentLog.Sections)-1]
			continue
		}

		// - Item
		if strings.HasPrefix(line, "- ") && currentSect != nil {
			text := strings.TrimPrefix(line, "- ")
			if currentSect.Title == "Docs" {
				if strings.HasPrefix(text, "Release") || strings.HasPrefix(text, "dev-") {
					continue
				}
			}
			currentSect.Items = append(currentSect.Items, ChangeItem{
				Text: text,
			})
		}
	}

	flushLog()

	return logs, scanner.Err()
}

func parseHeader(line string) (version, date string) {
	// ## [release-v1.0.3] - 2025-12-17
	line = strings.TrimPrefix(line, "## ")
	parts := strings.SplitN(line, " - ", 2)

	version = strings.Trim(parts[0], "[]")
	if len(parts) == 2 {
		date = parts[1]
	}
	return
}
