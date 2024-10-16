package utils

import (
	"strings"
	"unicode"
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

func SlugToTitle(input string) string {
	titled := strings.ToTitle(input)
	res := strings.ReplaceAll(titled, "-", " ")
	return res
}
