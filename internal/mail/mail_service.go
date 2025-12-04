package mail

import (
	"cchoice/internal/errs"
	"fmt"
	"strings"
)

//go:generate go tool stringer -type=MailService -trimprefix=MAIL_SERVICE_

type MailService int

const (
	MAIL_SERVICE_UNDEFINED MailService = iota
	MAIL_SERVICE_MAILEROO
)

func ParseMailServiceToEnum(ms string) MailService {
	switch strings.ToUpper(ms) {
	case MAIL_SERVICE_MAILEROO.String():
		return MAIL_SERVICE_MAILEROO
	default:
		panic(fmt.Errorf("%w: '%s'", errs.ErrCmdUndefinedService, ms))
	}
}
