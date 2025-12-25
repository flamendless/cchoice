package cmd

import (
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/mail"
	"cchoice/internal/mail/maileroo"
	"cchoice/internal/requests"
	"context"
	"fmt"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/gookit/goutil/dump"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"golang.org/x/sync/singleflight"
)

type testMailFlags struct {
	Service  string
	To       string
	CC       []string
	Subject  string
	Body     string
	Template string
}

var flagTestMail testMailFlags

func init() {
	f := cmdTestMail.Flags
	f().StringVarP(&flagTestMail.Service, "service", "s", "MAILEROO", "Mail service name")
	f().StringVarP(&flagTestMail.To, "to", "t", "", "Recipient email address")
	f().StringSliceVarP(&flagTestMail.CC, "cc", "c", nil, "CC email address(es)")
	f().StringVarP(&flagTestMail.Subject, "subject", "j", "Test Email", "Email subject")
	f().StringVarP(&flagTestMail.Body, "body", "b", "This is a test email from cchoice.", "Email body")
	f().StringVarP(&flagTestMail.Template, "template", "m", "", "Template file name (e.g., order_confirmation.html)")
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

		var errSendMail error
		if flagTestMail.Template != "" {
			ctx := context.Background()
			dbRO := database.New(database.DB_MODE_RO)
			settings, err := requests.GetSettingsData(
				ctx,
				fastcache.New(constants.CACHE_MAX_BYTES),
				&singleflight.Group{},
				dbRO,
				[]byte("test_mail_settings"),
				[]string{
					"mobile_no",
					"email",
					"address",
					"url_gmap",
					"url_waze",
					"url_facebook",
					"url_tiktok",
				},
			)
			if err != nil {
				panic(err)
			}
			cfg := conf.Conf()
			cfg.SetSettings(settings)

			data := mail.TemplateData{
				"LogoURL":          constants.PathEmailLogoCDN,
				"OrderNumber":      "CC-TEST-123456",
				"PaymentReference": "CCPM-ABC123DEF456",
				"LineItems": []map[string]any{
					{"Name": "Sample Product 1", "Quantity": 2, "Price": "₱1,000.00"},
					{"Name": "Sample Product 2", "Quantity": 1, "Price": "₱500.00"},
				},
				"Subtotal":        "₱2,500.00",
				"ShippingFee":     "₱150.00",
				"Total":           "₱2,650.00",
				"ShippingAddress": "123 Test Street, Barangay Test, Test City, Metro Manila 1234",
				"DeliveryETA":     "3-5 business days",
				"MobileNo":        cfg.Settings.MobileNo,
				"EMail":           cfg.Settings.EMail,
			}
			dump.Println(data)
			errSendMail = ms.SendTemplateEmail(flagTestMail.To, flagTestMail.CC, flagTestMail.Subject, flagTestMail.Template, data)
		} else {
			errSendMail = ms.SendEmail(flagTestMail.To, flagTestMail.CC, flagTestMail.Subject, flagTestMail.Body)
		}

		if errSendMail != nil {
			panic(errSendMail)
		}

		fmt.Println("Email sent successfully!")
	},
}
