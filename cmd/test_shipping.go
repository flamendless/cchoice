package cmd

import (
	"cchoice/internal/database"
	"cchoice/internal/errs"
	"cchoice/internal/geocoding/googlemaps"
	"cchoice/internal/logs"
	"cchoice/internal/shipping"
	"cchoice/internal/shipping/cchoice"
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
	f().StringVarP(&flagShippingService, "service", "s", "CCHOICE", "Shipping service name (LALAMOVE, CCHOICE)")
	rootCmd.AddCommand(cmdTestShipping)
}

var cmdTestShipping = &cobra.Command{
	Use:   "test_shipping",
	Short: "test shipping by making API calls",
	Run: func(cmd *cobra.Command, args []string) {
		var ss shipping.IShippingService
		switch flagShippingService {
		case shipping.SHIPPING_SERVICE_CCHOICE.String():
			ss = cchoice.MustInit()
		case shipping.SHIPPING_SERVICE_LALAMOVE.String():
			ss = lalamove.MustInit()
		default:
			panic(fmt.Errorf("%w: '%s'", errs.ErrCmdUnimplementedService, flagShippingService))
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

		geocoder := googlemaps.MustInit(nil)

		fmt.Printf("\nProvider Information:\n")
		fmt.Printf("    Provider: %s\n", capabilities.Provider)
		fmt.Printf("    API Version: %s\n", capabilities.APIVersion)
		fmt.Println()
		fmt.Println("=== Testing GetQuotation ===")
		reqShipping := shipping.ShippingRequest{
			PickupLocation: shipping.Location{
				Address: "Cavite, Philippines",
				Contact: shipping.Contact{
					Name:  "John Sender",
					Phone: "+639171234567",
				},
			},
			DeliveryLocation: shipping.Location{
				Address: "Quezon City, Philippines",
				Contact: shipping.Contact{
					Name:  "Jane Receiver",
					Phone: "+639179876543",
				},
			},
			Package: shipping.Package{
				Weight:      getPackageWeight(flagShippingService),
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
			ServiceType: shipping.SERVICE_TYPE_MOTORCYCLE,
			Options: map[string]any{
				"special_requests":   []string{"PURCHASE_SERVICE_1"},
				"language":           "en_PH",
				"is_route_optimized": true,
			},
		}

		fmt.Printf("\nUse geocoding to get coordinates for both locations\n")
		pickupCoords, err := geocoder.GeocodeShippingAddress(reqShipping.PickupLocation.Address)
		if err != nil {
			panic(err)
		}
		reqShipping.PickupLocation.Coordinates = shipping.Coordinates{
			Lat: pickupCoords.Lat,
			Lng: pickupCoords.Lng,
		}

		deliveryCoords, err := geocoder.GeocodeShippingAddress(reqShipping.DeliveryLocation.Address)
		if err != nil {
			panic(err)
		}
		reqShipping.DeliveryLocation.Coordinates = shipping.Coordinates{
			Lat: deliveryCoords.Lat,
			Lng: deliveryCoords.Lng,
		}

		fmt.Println("Quotation Request:", reqShipping)

		quotation, err := ss.GetQuotation(reqShipping)

		db := database.New(database.DB_MODE_RW)
		logs.LogExternalAPICall(cmd.Context(), db.GetQueries(), logs.ExternalAPILogParams{
			CheckoutID: nil,
			Service:    "shipping",
			API:        ss.Enum(),
			Endpoint:   "/v3/quotations",
			HTTPMethod: "POST",
			Payload:    reqShipping,
			Response:   quotation,
			Error:      err,
		})

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

		if flagShippingService == shipping.SHIPPING_SERVICE_CCHOICE.String() {
			fmt.Println("\n=== C-Choice Service: Testing Unsupported Operations (Should Fail) ===")

			fmt.Println("Testing CreateOrder (should fail)...")
			_, err = ss.CreateOrder(reqShipping)
			if err != nil {
				fmt.Printf("CreateOrder failed as expected: %v\n", err)
			} else {
				fmt.Println("CreateOrder should have failed!")
			}

			fmt.Println("Testing GetOrderStatus (should fail)...")
			_, err = ss.GetOrderStatus("test-order-id")
			if err != nil {
				fmt.Printf("GetOrderStatus failed as expected: %v\n", err)
			} else {
				fmt.Println("GetOrderStatus should have failed!")
			}

			fmt.Println("Testing CancelOrder (should fail)...")
			err = ss.CancelOrder("test-order-id")
			if err != nil {
				fmt.Printf("CancelOrder failed as expected: %v\n", err)
			} else {
				fmt.Println("CancelOrder should have failed!")
			}

			fmt.Println("\n=== Test Completed ===")
			fmt.Println("C-Choice shipping service (quotation-only) working successfully!")
			return
		}

		fmt.Println("=== Testing CreateOrder ===")

		orderParams := lalamove.OrderRequestParams{
			IsPODEnabled: true,
			Partner:      "Lalamove Partner 1",
			Remarks:      "Please handle with care - fragile items",
			Metadata: map[string]string{
				"restaurant_order_id": "1234",
				"restaurant_name":     "Rustam's Kebab",
				"customer_notes":      "Extra spicy",
				"order_source":        "mobile_app",
			},
		}

		orderReq := lalamove.CreateOrderRequest(reqShipping, quotation, orderParams)

		fmt.Println("Order Request (with additional fields):")
		fmt.Printf("  Quotation ID: %v\n", orderReq.Options["quotation_id"])
		fmt.Printf("  POD Enabled: %v\n", orderReq.Options["is_pod_enabled"])
		fmt.Printf("  Partner: %v\n", orderReq.Options["partner"])
		fmt.Printf("  Remarks: %v\n", orderReq.Options["remarks"])
		dump.Println("  Metadata:", orderReq.Options["metadata"])

		order, err := ss.CreateOrder(orderReq)

		logs.LogExternalAPICall(cmd.Context(), db.GetQueries(), logs.ExternalAPILogParams{
			CheckoutID: nil,
			Service:    "shipping",
			API:        ss.Enum(),
			Endpoint:   "/v3/orders",
			HTTPMethod: "POST",
			Payload:    orderReq,
			Response:   order,
			Error:      err,
		})

		if err != nil {
			panic(err)
		}

		fmt.Printf("Order ID: %s\n", order.ID)
		fmt.Printf("Order Status: %s\n", order.Status)
		dump.Println("Order Metadata:", order.Metadata)

		fmt.Println("=== Testing GetOrderStatus ===")
		orderStatus, err := ss.GetOrderStatus(order.ID)

		logs.LogExternalAPICall(cmd.Context(), db.GetQueries(), logs.ExternalAPILogParams{
			CheckoutID: nil,
			Service:    "shipping",
			API:        ss.Enum(),
			Endpoint:   "/v3/orders/" + order.ID,
			HTTPMethod: "GET",
			Payload:    map[string]string{"order_id": order.ID},
			Response:   orderStatus,
			Error:      err,
		})

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

func getPackageWeight(service string) string {
	switch service {
	case shipping.SHIPPING_SERVICE_CCHOICE.String():
		return "2.5"
	case shipping.SHIPPING_SERVICE_LALAMOVE.String():
		return "LESS_THAN_3KG"
	default:
		return "2.5"
	}
}
