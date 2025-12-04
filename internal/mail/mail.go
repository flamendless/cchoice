package mail

type EmailMessage struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

type TemplateData map[string]any

type IMailService interface {
	Enum() MailService
	SendEmail(to []string, subject, body string) error
	SendTemplateEmail(to []string, subject, templateName string, data TemplateData) error
}
