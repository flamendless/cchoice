package utils

import "unicode"

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
