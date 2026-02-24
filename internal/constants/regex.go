package constants

import "regexp"

var (
	SizeRegex           = regexp.MustCompile(`/\d+x\d+/`)
	MultipleSpacesRegex = regexp.MustCompile(`\s+`)
	OrderReferenceRegex = regexp.MustCompile(`^CCO-[0-9a-zA-Z]+[0-9A-F]{6}$`)
	PasswordRegex       = regexp.MustCompile(`^[a-zA-Z0-9\-_.?#@]{1,64}$`)
	EmailRegex          = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func ToPath1280(path string) string {
	return SizeRegex.ReplaceAllString(path, "/1280x1280/")
}
