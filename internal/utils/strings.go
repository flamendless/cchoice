package utils

import (
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var caserPool = sync.Pool{
	New: func() any { return cases.Title(language.English) },
}

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
	caser := caserPool.Get().(cases.Caser)
	keywords := strings.Split(input, "-")
	res := strings.Join(keywords, " ")
	res = caser.String(res)
	caserPool.Put(caser)
	return res
}
