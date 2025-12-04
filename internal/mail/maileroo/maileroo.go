package maileroo

import (
	"bytes"
	"cchoice/internal/conf"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/mail"
	"context"
	"errors"
	"html/template"
	"os"
	"path/filepath"

	"github.com/maileroo/maileroo-go-sdk/maileroo"
	"go.uber.org/zap"
)

const defaultTimeoutSeconds = 30

type Maileroo struct {
	client      *maileroo.Client
	from        maileroo.EmailAddress
	mailService mail.MailService
}

func validate() {
	cfg := conf.Conf()
	if cfg.MailService != mail.MAIL_SERVICE_MAILEROO.String() {
		panic(errs.ErrMailerooServiceInit)
	}
	if cfg.MailerooConfig.APIKey == "" {
		panic(errs.ErrMailerooAPIKeyRequired)
	}
	if cfg.MailerooConfig.From == "" {
		panic(errs.ErrMailerooFromRequired)
	}
}

func MustInit() *Maileroo {
	validate()

	cfg := conf.Conf()
	client, err := maileroo.NewClient(cfg.MailerooConfig.APIKey, defaultTimeoutSeconds)
	if err != nil {
		panic(errors.Join(errs.ErrMailerooServiceInit, err))
	}

	return &Maileroo{
		mailService: mail.MAIL_SERVICE_MAILEROO,
		client:      client,
		from:        maileroo.NewEmail(cfg.MailerooConfig.From, "C-Choice Construction Supply Shop"),
	}
}

func (m *Maileroo) Enum() mail.MailService {
	return m.mailService
}

func (m *Maileroo) SendEmail(to []string, subject, body string) error {
	return m.sendEmail(context.Background(), to, subject, body, false)
}

func (m *Maileroo) SendTemplateEmail(to []string, subject, templateName string, data mail.TemplateData) error {
	templatePath := filepath.Join("templates", templateName)
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return errors.Join(errs.ErrTemplateRead, err)
	}

	tmpl, err := template.New(templateName).Parse(string(tmplContent))
	if err != nil {
		return errors.Join(errs.ErrTemplateParse, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return errors.Join(errs.ErrTemplateExecute, err)
	}

	return m.sendEmail(context.Background(), to, subject, buf.String(), true)
}

func (m *Maileroo) sendEmail(ctx context.Context, to []string, subject, body string, isHTML bool) error {
	const logTag = "[Maileroo Send Email]"

	toAddresses := make([]maileroo.EmailAddress, 0, len(to))
	for _, addr := range to {
		toAddresses = append(toAddresses, maileroo.NewEmail(addr, ""))
	}

	emailData := maileroo.BasicEmailData{
		From:    m.from,
		To:      toAddresses,
		Subject: subject,
	}

	if isHTML {
		emailData.HTML = maileroo.StrPtr(body)
	} else {
		emailData.Plain = maileroo.StrPtr(body)
	}

	referenceID, err := m.client.SendBasicEmail(ctx, emailData)
	if err != nil {
		logs.Log().Error(logTag, zap.Error(err))
		return errors.Join(errs.ErrMailerooSendFailed, err)
	}

	logs.Log().Info(
		logTag,
		zap.Strings("to", to),
		zap.String("subject", subject),
		zap.String("reference_id", referenceID),
	)

	return nil
}

var _ mail.IMailService = (*Maileroo)(nil)
