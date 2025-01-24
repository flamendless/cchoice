package utils

import (
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var caserOnce sync.Once
var caser cases.Caser

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
	caserOnce.Do(func() {
		caser = cases.Title(language.English)
	})
	keywords := strings.Split(input, "-")
	res := strings.Join(keywords, " ")
	res = caser.String(res)
	return res
}
