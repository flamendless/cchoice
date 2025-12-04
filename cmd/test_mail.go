package cmd

import (
	"cchoice/internal/mail"
	"cchoice/internal/mail/maileroo"
	"fmt"

	"github.com/gookit/goutil/dump"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
)

type testMailFlags struct {
	Service string
	To      []string
	Subject string
	Body    string
}

var flagTestMail testMailFlags

func init() {
	f := cmdTestMail.Flags
	f().StringVarP(&flagTestMail.Service, "service", "s", "MAILEROO", "Mail service name")
	f().StringSliceVarP(&flagTestMail.To, "to", "t", nil, "Recipient email address(es)")
	f().StringVarP(&flagTestMail.Subject, "subject", "j", "Test Email", "Email subject")
	f().StringVarP(&flagTestMail.Body, "body", "b", "This is a test email from cchoice.", "Email body")
	if err := cmdTestMail.MarkFlagRequired("to"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(cmdTestMail)
}

var cmdTestMail = &cobra.Command{
	Use:   "test_mail",
	Short: "test mail by sending a test email",
	Run: func(cmd *cobra.Command, args []string) {
		var ms mail.IMailService
		switch flagTestMail.Service {
		case mail.MAIL_SERVICE_MAILEROO.String():
			ms = maileroo.MustInit()
		default:
			panic(fmt.Sprintf("Unimplemented mail service: '%s'", flagTestMail.Service))
		}

		dump.Println(flagTestMail)

		if err := ms.SendEmail(flagTestMail.To, flagTestMail.Subject, flagTestMail.Body); err != nil {
			panic(err)
		}

		fmt.Println("Email sent successfully!")
	},
}
