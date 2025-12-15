package components

import "cchoice/internal/conf"

const placeholderImageSrc = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='256' height='256'%3E%3C/svg%3E"

func URL(path string) string {
	if conf.Conf().IsProd() {
		return path
	}
	return "/cchoice" + path
}
