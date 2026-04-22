package models

import "github.com/a-h/templ"

type Platform struct {
	Label string
	Value string
	Icon  templ.Component
	Order int
}
