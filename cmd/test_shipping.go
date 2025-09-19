package cmd

import (
	"cchoice/internal/shipping"
	"cchoice/internal/shipping/lalamove"
	"fmt"
	"time"

	"github.com/gookit/goutil/dump"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
)

var flagShippingService string

func init() {
	f := cmdTestShipping.Flags
	f().StringVarP(&flagShippingService, "service", "s", "lalamove", "Shipping service name")
	rootCmd.AddCommand(cmdTestShipping)
}

var cmdTestShipping = &cobra.Command{
	Use:   "test_shipping",
	Short: "test shipping by making API calls",
	Run: func(cmd *cobra.Command, args []string) {
		var ss shipping.IShippingService
		switch flagShippingService {
		case "lalamove":
			ss = lalamove.MustInit()
		default:
			panic(fmt.Sprintf("Unimplemented shipping service: '%s'", flagShippingService))
		}

		fmt.Println("=== Testing GetCapabilities ===")
		capabilities, err := ss.GetCapabilities()
		if err != nil {
			panic(err)
		}
		fmt.Println("Supported Services:", capabilities.SupportedServices)
		fmt.Println("Coverage Areas:", capabilities.Coverage)
		fmt.Println("Features:")
		fmt.Printf("    Real Time Tracking: %v\n", capabilities.Features.RealTimeTracking)
		fmt.Printf("    Route Optimization: %v\n", capabilities.Features.RouteOptimization)
		fmt.Printf("    Scheduled Delivery: %v\n", capabilities.Features.ScheduledDelivery)
		fmt.Printf("    Special Requests: %v\n", capabilities.Features.SpecialRequests)
		fmt.Printf("    Multiple Stops: %v\n", capabilities.Features.MultipleStops)
		fmt.Printf("    Weight Based Pricing: %v\n", capabilities.Features.WeightBasedPricing)
		fmt.Printf("    Insurance: %v\n", capabilities.Features.Insurance)
		fmt.Printf("    Proof of Delivery: %v\n", capabilities.Features.ProofOfDelivery)
		fmt.Printf("    Cash on Delivery: %v\n", capabilities.Features.CashOnDelivery)
		fmt.Printf("    Contactless Delivery: %v\n", capabilities.Features.ContactlessDelivery)

		fmt.Printf("\nProvider Information:\n")
		fmt.Printf("    Provider: %s\n", capabilities.Provider)
		fmt.Printf("    API Version: %s\n", capabilities.APIVersion)
		fmt.Println()
		fmt.Println("=== Testing GetQuotation ===")
		quotationReq := shipping.ShippingRequest{
			PickupLocation: shipping.Location{
				Coordinates: shipping.Coordinates{
					Lat: "14.4791",
					Lng: "120.8970",
				},
				Address: "Cavite, Philippines",
				Contact: shipping.Contact{
					Name:  "John Sender",
					Phone: "+63-917-123-4567",
				},
			},
			DeliveryLocation: shipping.Location{
				Coordinates: shipping.Coordinates{
					Lat: "14.6760",
					Lng: "121.0437",
				},
				Address: "Quezon City, Philippines",
				Contact: shipping.Contact{
					Name:  "Jane Receiver",
					Phone: "+63-917-987-6543",
				},
			},
			Package: shipping.Package{
				Weight:      "LESS_THAN_3KG",
				Description: "Electronics and office supplies",
				Value:       "5000",
				Dimensions: map[string]string{
					"length": "30",
					"width":  "20",
					"height": "15",
				},
				Metadata: map[string]any{
					"categories":            []string{"FOOD_DELIVERY", "OFFICE_ITEM"},
					"quantity":              "3",
					"handling_instructions": []string{"KEEP_UPRIGHT"},
				},
			},
			ScheduledAt: time.Now().UTC().Add(1 * time.Hour).Format("2006-01-02T15:04:05Z"),
			ServiceType: "MOTORCYCLE",
			Options: map[string]any{
				"special_requests":   []string{"PURCHASE_SERVICE_1"},
				"language":           "en_PH",
				"is_route_optimized": true,
			},
		}
		fmt.Println("Quotation Request:", quotationReq)

		quotation, err := ss.GetQuotation(quotationReq)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Quotation ID: %s\n", quotation.ID)
		fmt.Printf("Fee: %s %.2f\n", quotation.Currency, quotation.Fee)
		fmt.Printf("Distance: %.2f km\n", quotation.DistanceKm)
		fmt.Printf("Service Type: %s\n", quotation.ServiceType)
		fmt.Printf("Expires At: %s\n", quotation.ExpiresAt)
		if quotation.EstimatedETA > 0 {
			fmt.Printf("Estimated ETA: %d minutes\n", quotation.EstimatedETA)
		}
		dump.Println("Quotation Metadata:", quotation.Metadata)

		fmt.Println("=== Testing CreateOrder ===")
		orderReq := shipping.ShippingRequest{
			PickupLocation: shipping.Location{
				Coordinates: shipping.Coordinates{
					Lat: "14.5995",
					Lng: "120.9842",
				},
				Address: "Makati, Philippines",
				Contact: shipping.Contact{
					Name:  "Business Sender",
					Phone: "+63-917-111-2222",
				},
			},
			DeliveryLocation: shipping.Location{
				Coordinates: shipping.Coordinates{
					Lat: "14.6760",
					Lng: "121.0437",
				},
				Address: "Ortigas, Philippines",
				Contact: shipping.Contact{
					Name:  "Customer Receiver",
					Phone: "+63-917-333-4444",
				},
			},
			Package: shipping.Package{
				Weight:      "LESS_THAN_3KG",
				Description: "Food delivery items",
				Value:       "1500",
				Metadata: map[string]any{
					"categories":            []string{"FOOD_DELIVERY"},
					"quantity":              "2",
					"handling_instructions": []string{"KEEP_UPRIGHT", "FRAGILE"},
				},
			},
			ScheduledAt: time.Now().UTC().Add(2 * time.Hour).Format("2006-01-02T15:04:05Z"),
			ServiceType: "MOTORCYCLE",
			Options: map[string]any{
				"special_requests":   []string{"CASH_ON_DELIVERY"},
				"language":           "en_PH",
				"is_route_optimized": true,
			},
		}

		order, err := ss.CreateOrder(orderReq)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Order ID: %s\n", order.ID)
		fmt.Printf("Order Status: %s\n", order.Status)
		dump.Println("Order Metadata:", order.Metadata)

		fmt.Println("=== Testing GetOrderStatus ===")
		orderStatus, err := ss.GetOrderStatus(order.ID)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Order Status ID: %s\n", orderStatus.ID)
		fmt.Printf("Current Status: %s\n", orderStatus.Status)
		if orderStatus.Quotation.Fee > 0 {
			fmt.Printf("Order Fee: %s %.2f\n", orderStatus.Quotation.Currency, orderStatus.Quotation.Fee)
			fmt.Printf("Order Distance: %.2f km\n", orderStatus.Quotation.DistanceKm)
			if orderStatus.Quotation.EstimatedETA > 0 {
				fmt.Printf("ETA: %d minutes\n", orderStatus.Quotation.EstimatedETA)
			}
		}
		if len(orderStatus.TrackingInfo) > 0 {
			dump.Println("Tracking Info:", orderStatus.TrackingInfo)
		}

		fmt.Println("=== Test Completed ===")
		fmt.Println("Generic shipping interface working successfully!")
	},
}
