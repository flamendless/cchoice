package utils

import (
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
