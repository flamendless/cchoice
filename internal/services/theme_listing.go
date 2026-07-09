package services

import "cchoice/internal/database/queries"

type themeListingRow struct {
	tblTheme queries.TblTheme
	active   bool
}

func normalizeThemeListingSort(sortBy, sortDir string) (string, string) {
	switch sortBy {
	case "TITLE", "START_DATE", "END_DATE", "STATUS":
	default:
		sortBy = "TITLE"
	}
	switch sortDir {
	case "ASC", "DESC":
	default:
		sortDir = "ASC"
	}
	return sortBy, sortDir
}

func parseThemeActive(active any) bool {
	switch v := active.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case int:
		return v != 0
	case []byte:
		return string(v) == "1"
	case string:
		return v == "1"
	default:
		return false
	}
}

func mapThemeRowsFromTitleAsc(rows []queries.SearchThemesSortTitleAscRow) []themeListingRow {
	out := make([]themeListingRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, themeListingRow{tblTheme: r.TblTheme, active: parseThemeActive(r.Active)})
	}
	return out
}

func mapThemeRowsFromTitleDesc(rows []queries.SearchThemesSortTitleDescRow) []themeListingRow {
	out := make([]themeListingRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, themeListingRow{tblTheme: r.TblTheme, active: parseThemeActive(r.Active)})
	}
	return out
}

func mapThemeRowsFromStartDateAsc(rows []queries.SearchThemesSortStartDateAscRow) []themeListingRow {
	out := make([]themeListingRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, themeListingRow{tblTheme: r.TblTheme, active: parseThemeActive(r.Active)})
	}
	return out
}

func mapThemeRowsFromStartDateDesc(rows []queries.SearchThemesSortStartDateDescRow) []themeListingRow {
	out := make([]themeListingRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, themeListingRow{tblTheme: r.TblTheme, active: parseThemeActive(r.Active)})
	}
	return out
}

func mapThemeRowsFromEndDateAsc(rows []queries.SearchThemesSortEndDateAscRow) []themeListingRow {
	out := make([]themeListingRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, themeListingRow{tblTheme: r.TblTheme, active: parseThemeActive(r.Active)})
	}
	return out
}

func mapThemeRowsFromEndDateDesc(rows []queries.SearchThemesSortEndDateDescRow) []themeListingRow {
	out := make([]themeListingRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, themeListingRow{tblTheme: r.TblTheme, active: parseThemeActive(r.Active)})
	}
	return out
}

func mapThemeRowsFromStatusAsc(rows []queries.SearchThemesSortStatusAscRow) []themeListingRow {
	out := make([]themeListingRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, themeListingRow{tblTheme: r.TblTheme, active: parseThemeActive(r.Active)})
	}
	return out
}

func mapThemeRowsFromStatusDesc(rows []queries.SearchThemesSortStatusDescRow) []themeListingRow {
	out := make([]themeListingRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, themeListingRow{tblTheme: r.TblTheme, active: parseThemeActive(r.Active)})
	}
	return out
}
