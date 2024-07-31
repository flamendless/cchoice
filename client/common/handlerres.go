package common

import "github.com/a-h/templ"

type HandlerRes struct {
	Component  templ.Component
	Error      error
	StatusCode int
	RedirectTo string
	ReplaceURL string
}

type AuthSession struct {
	Token   string
	NeedOTP bool
}

type User struct {
	ID         string
	FirstName  string
	MiddleName string
	LastName   string
	Email      string
}
