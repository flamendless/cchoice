package common

import "github.com/a-h/templ"

type FooterDetails struct {
	URLTikTok   templ.SafeURL
	URLFacebook templ.SafeURL
	URLGMap     templ.SafeURL
	Email       templ.SafeURL
	MobileNo    templ.SafeURL
}
