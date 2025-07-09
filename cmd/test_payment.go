package cmd

import (
	"cchoice/internal/database"
	"cchoice/internal/payments"
	"cchoice/internal/payments/paymongo"
	"fmt"

	"github.com/Rhymond/go-money"
	"github.com/gookit/goutil/dump"
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
		var pg payments.IPaymentGateway
		switch flagGateway {
		case "paymongo":
			pg = paymongo.MustInit()
		default:
			panic(fmt.Sprintf("Unimplemented gateway: '%s'", flagGateway))
		}

		resPaymentMethods, err := pg.GetAvailablePaymentMethods()
		if err != nil {
			panic(err)
		}
		dump.Println("Available payment methods", resPaymentMethods)

		payload := paymongo.CreateCheckoutSessionPayload{
			Data: paymongo.CreateCheckoutSessionData{
				Attributes: paymongo.CreateCheckoutSessionAttr{
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
					Description:         "test description",
					PaymentMethodTypes:  resPaymentMethods.ToPaymentMethods(),
					ReferenceNumber:     "test-ref-number",
					SendEmailReceipt:    false,
					ShowDescription:     true,
					ShowLineItems:       true,
					StatementDescriptor: "test statement descriptor",
				},
			},
		}
		resCheckout, err := pg.CreateCheckoutSession(&payload)
		if err != nil {
			panic(err)
		}
		dump.Println("Checkout", resCheckout)

		db := database.New(database.DB_MODE_RW)
		inserted, err := db.GetQueries().CreateCheckout(cmd.Context(), *resCheckout.ToCheckout(pg))
		if err != nil {
			panic(err)
		}
		dump.Println("Inserted checkout", inserted)

		for _, lineItem := range resCheckout.ToLineItems() {
			inserted, err := db.GetQueries().CreateCheckoutLineItem(cmd.Context(), *lineItem)
			if err != nil {
				panic(err)
			}
			dump.Println("Inserted line item", inserted)
		}
	},
}
