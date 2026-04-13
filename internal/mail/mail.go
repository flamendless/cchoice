package mail

type EmailMessage struct {
	To      string
	Subject string
	Body    string
	CC      []string
	IsHTML  bool
}

type TemplateData map[string]any

type IMailService interface {
	Enum() MailService
	SendEmail(to string, cc []string, subject, body string) error
	SendTemplateEmail(to string, cc []string, subject, templateName string, data TemplateData) error
}
