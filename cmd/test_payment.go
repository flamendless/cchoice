package cmd

import (
	"cchoice/internal/enums"
	"cchoice/internal/payments"
	"fmt"

	"github.com/Rhymond/go-money"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
)

var flagGateway string

func init() {
	f := cmdTestPayment.Flags
	f().StringVarP(&flagGateway, "gateway", "g", "paymongo", "Gateway name")
	rootCmd.AddCommand(cmdTestPayment)
}

var cmdTestPayment = &cobra.Command{
	Use:   "test_payment",
	Short: "test_payment",
	Run: func(cmd *cobra.Command, args []string) {
		var pg payments.IPayments
		switch flagGateway {
		case "paymongo":
			pg = payments.MustInitPayMongo()
		default:
			panic(fmt.Sprintf("Unimplemented gateway: '%s'", flagGateway))
		}

		payload := payments.PayMongoCreateCheckoutSessionPayload{
			Data: payments.PayMongoCreateCheckoutSessionData{
				Attributes: payments.PayMongoCreateCheckoutSessionAttr{
					CancelURL:  "https://test.com/cancel",
					SuccessURL: "https://test.com/success",
					Billing: payments.Billing{
						Address: payments.Address{
							Line1:      "test line 1",
							Line2:      "test line 2",
							City:       "test city",
							State:      "test state",
							PostalCode: "test postal code",
							Country:    "PH",
						},
						Name:  "test name",
						Email: "test@mail.com",
						Phone: "test phone",
					},
					LineItems: []payments.LineItem{
						{
							Amount:      1000,
							Currency:    money.PHP,
							Description: "test line item description",
							Images:      []string{"https://test.com/image"},
							Name:        "test line item name",
							Quantity:    2,
						},
					},
					Description: "test description",
					PaymentMethodTypes: []enums.PaymentMethod{
						enums.PAYMENT_METHOD_QRPH,
					},
					ReferenceNumber:     "test-ref-number",
					SendEmailReceipt:    false,
					ShowDescription:     true,
					ShowLineItems:       true,
					StatementDescriptor: "test statement descriptor",
				},
			},
		}
		res, err := pg.CreateCheckoutSession(&payload)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)
	},
}
