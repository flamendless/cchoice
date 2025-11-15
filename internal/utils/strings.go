package utils

import (
	"cchoice/internal/enums"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetInitials(str string) string {
	res := make([]rune, 0, len(str))
	for _, c := range str {
		if unicode.IsSpace(c) || unicode.IsSymbol(c) {
			continue
		}
		if unicode.IsUpper(c) {
			res = append(res, c)
		}
	}
	return string(res)
}

func RemoveEmptyStrings(input []string) []string {
	var res []string
	for _, val := range input {
		if val != "" {
			res = append(res, val)
		}
	}
	return res
}

func SlugToTile(input string) string {
	caser := cases.Title(language.English)
	keywords := strings.Split(input, "-")
	joined := strings.Join(keywords, " ")
	return caser.String(joined)
}

func GetBoolFlag(flag string) bool {
	return flag == "true" || flag == "1"
}

func LabelToID(module enums.Module, label string) string {
	result := strings.ToLower(label)
	result = strings.ReplaceAll(result, " ", "-")
	var builder strings.Builder
	for _, r := range result {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			builder.WriteRune(r)
		}
	}
	moduleStr := strings.ToLower(module.String())
	return moduleStr + "-" + builder.String()
}
