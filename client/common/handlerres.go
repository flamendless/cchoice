package common

import "github.com/a-h/templ"

type HandlerRes struct {
	Component  templ.Component
	Error      error
	StatusCode int
}
