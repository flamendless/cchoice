package common

import "github.com/a-h/templ"

func ThemeStyle(themeCSS string) templ.Component {
	if themeCSS == "" {
		return templ.NopComponent
	}
	return templ.Raw("<style>" + themeCSS + "</style>")
}
