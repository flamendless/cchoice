package constants

import "regexp"

const (
	Pattern1280           = `/1280x1280/`
	PatternPassword       = `[a-zA-Z0-9\-_.?#@]{8,32}`
	PatternEmail          = `[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`
	PatternName           = `[a-zA-Z\s\-]+`
	PatternMobileNumberFE = `[0-9]{10}`
	PatternMobileNumber   = `\+63[0-9]{10}`
	PatternOrderReference = `CCO-[0-9a-zA-Z]+[0-9A-F]{6}`
	PatternPostalCode     = `[0-9]{4}`
	PatternSize           = `/\d+x\d+/`
	PatternSubcategory    = `[a-zA-Z0-9\s\-_]+`
	PatternProductName    = `[a-zA-Z0-9\s\-_.,]+`
	PatternSerialNumber   = `[a-zA-Z0-9\-_]+`
	PatternMultipleSpaces = `\s+`
)

var (
	ReSize           = regexp.MustCompile(PatternSize)
	ReMultipleSpaces = regexp.MustCompile(PatternMultipleSpaces)
	ReOrderReference = regexp.MustCompile(`^` + PatternOrderReference + `$`)
	RePassword       = regexp.MustCompile(`^` + PatternPassword + `$`)
	ReEmail          = regexp.MustCompile(`^` + PatternEmail + `$`)
	ReName           = regexp.MustCompile(`^` + PatternName + `$`)
	ReMobileNumber   = regexp.MustCompile(`^` + PatternMobileNumber + `$`)
	RePostalCode     = regexp.MustCompile(`^` + PatternPostalCode + `$`)
)

func ToPath1280(path string) string {
	return ReSize.ReplaceAllString(path, Pattern1280)
}
