package cmd

import (
	"cchoice/internal/payments/paymongo"
	"fmt"

	"github.com/gookit/goutil/dump"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
)

type createPaymongoWebhooksFlags struct {
	LiveMode bool
	URL      string
	Events   []string
}

var flagCreatePaymongoWebhooks createPaymongoWebhooksFlags

func init() {
	f := cmdCreatePaymongoWebhooks.Flags
	f().BoolVarP(&flagCreatePaymongoWebhooks.LiveMode, "live", "l", false, "Use live mode instead of test mode")
	f().StringVarP(&flagCreatePaymongoWebhooks.URL, "url", "u", "", "Webhook URL endpoint (e.g., https://your-domain.com/cchoice/webhooks/paymongo)")
	f().StringSliceVarP(&flagCreatePaymongoWebhooks.Events, "events", "e", nil, "Webhook events to subscribe to (default: payment.paid, payment.failed, checkout_session.payment.paid)")
	if err := cmdCreatePaymongoWebhooks.MarkFlagRequired("url"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(cmdCreatePaymongoWebhooks)
}

var cmdCreatePaymongoWebhooks = &cobra.Command{
	Use:   "create_paymongo_webhooks",
	Short: "Register webhooks with PayMongo",
	Long:  "Register webhooks with PayMongo payment gateway",
	Run: func(cmd *cobra.Command, args []string) {
		pg := paymongo.MustInit()

		events := flagCreatePaymongoWebhooks.Events
		if len(events) == 0 {
			events = []string{
				paymongo.WebhookEventPaymentPaid,
				paymongo.WebhookEventPaymentFailed,
				paymongo.WebhookEventCheckoutSessionPaymentPaid,
			}
		}

		fmt.Println("Creating PayMongo webhook...")
		fmt.Printf("  Mode: %s\n", modeString(flagCreatePaymongoWebhooks.LiveMode))
		fmt.Printf("  URL: %s\n", flagCreatePaymongoWebhooks.URL)
		fmt.Printf("  Events: %v\n", events)

		result, err := pg.CreateWebhook(flagCreatePaymongoWebhooks.URL, events)
		if err != nil {
			panic(fmt.Sprintf("Failed to create webhook: %v", err))
		}

		fmt.Println("\nWebhook created successfully!")
		fmt.Printf("  Webhook ID: %s\n", result.Data.ID)
		fmt.Printf("  Status: %s\n", result.Data.Attributes.Status)
		fmt.Printf("  Live Mode: %v\n", result.Data.Attributes.Livemode)
		fmt.Printf("  URL: %s\n", result.Data.Attributes.URL)
		fmt.Printf("  Events: %v\n", result.Data.Attributes.Events)

		fmt.Println("\n=== IMPORTANT ===")
		fmt.Println("Save the following secret key in your environment variables:")
		fmt.Printf("  PAYMONGO_WEBHOOK_SECRET_KEY=%s\n", result.Data.Attributes.SecretKey)
		fmt.Println("\nThis secret key is required to verify webhook signatures.")
		fmt.Println("You will not be able to retrieve this key again from PayMongo.")

		if cmd.Flags().Changed("verbose") {
			dump.Println("Full response:", result)
		}
	},
}

func modeString(isLive bool) string {
	if isLive {
		return "live"
	}
	return "test"
}
