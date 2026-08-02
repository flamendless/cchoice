package mail

type EmailMessage struct {
	To      string
	Subject string
	Body    string
	CC      []string
	IsHTML  bool
}

type TemplateData map[string]any

type Attachment struct {
	FileName    string
	ContentType string
	Content     []byte
}

type IMailService interface {
	Enum() MailService
	SendEmail(to string, cc []string, subject, body string) error
	SendTemplateEmail(to string, cc []string, subject, templateName string, data TemplateData) error
	SendTemplateEmailWithAttachments(to string, cc []string, subject, templateName string, data TemplateData, attachments []Attachment) error
}
