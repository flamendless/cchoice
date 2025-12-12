package constants

import "regexp"

var (
	SizeRegex           = regexp.MustCompile(`/\d+x\d+/`)
	MultipleSpacesRegex = regexp.MustCompile(`\s+`)
	OrderReferenceRegex = regexp.MustCompile(`^CCO-[0-9a-zA-Z]+[0-9A-F]{6}$`)
)

func ToPath1280(path string) string {
	return SizeRegex.ReplaceAllString(path, "/1280x1280/")
}
