package cmd

import (
	"cchoice/internal/database"
	"cchoice/internal/errs"
	"cchoice/internal/geocoding/googlemaps"
	"cchoice/internal/logs"
	"cchoice/internal/shipping"
	"cchoice/internal/shipping/cchoice"
	"cchoice/internal/shipping/lalamove"
	"context"
	"fmt"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
)

var flagShippingService string

type TestLocation struct {
	Name    string
	Address string
	State   string
}

type TestCase struct {
	Location TestLocation
	Weight   string
}

type QuotationResult struct {
	Destination  string
	Weight       string
	Fee          float64
	Distance     float64
	ETA          int
	FreeDelivery bool
	Error        error
}

func init() {
	f := cmdTestShipping.Flags
	f().StringVarP(&flagShippingService, "service", "s", "CCHOICE", "Shipping service name (LALAMOVE, CCHOICE)")
	rootCmd.AddCommand(cmdTestShipping)
}

var cmdTestShipping = &cobra.Command{
	Use:   "test_shipping",
	Short: "test shipping by making API calls",
	Run: func(cmd *cobra.Command, args []string) {
		switch flagShippingService {
		case shipping.SHIPPING_SERVICE_CCHOICE.String():
			testCChoiceService(cmd.Context())
		case shipping.SHIPPING_SERVICE_LALAMOVE.String():
			testLalamoveService(cmd.Context())
		default:
			panic(fmt.Errorf("%w: '%s'", errs.ErrCmdUnimplementedService, flagShippingService))
		}
	},
}

func testCChoiceService(ctx context.Context) {
	ss := cchoice.MustInit()
	db := database.New(database.DB_MODE_RW)
	geocoder := googlemaps.MustInit(db)

	fmt.Println("=== C-Choice Shipping Service - Matrix Test ===")
	fmt.Println()

	pickupLocation := TestLocation{
		Name:    "General Trias, Cavite",
		Address: "General Trias, Cavite, Philippines",
		State:   "Cavite",
	}
	deliveryLocations := []TestLocation{
		{Name: "Imus, Cavite", Address: "Imus, Cavite, Philippines", State: "Cavite"},
		{Name: "Pasay City", Address: "Pasay City, Metro Manila, Philippines", State: "Metro Manila"},
		{Name: "Quezon City", Address: "Quezon City, Metro Manila, Philippines", State: "Metro Manila"},
	}

	weights := []string{"0.5", "8.0", "18.0"}
	contact := shipping.Contact{
		Name:  "Juan Dela Cruz",
		Phone: "+639171234567",
	}

	pickupCoords, err := geocoder.GeocodeShippingAddress(pickupLocation.Address)
	if err != nil {
		panic(err)
	}

	results := []QuotationResult{}
	for _, location := range deliveryLocations {
		deliveryCoords, err := geocoder.GeocodeShippingAddress(location.Address)
		if err != nil {
			fmt.Printf("Error geocoding %s: %v\n", location.Name, err)
			continue
		}

		for _, weight := range weights {
			req := shipping.ShippingRequest{
				PickupLocation: shipping.Location{
					Address: pickupLocation.Address,
					Coordinates: shipping.Coordinates{
						Lat: pickupCoords.Lat,
						Lng: pickupCoords.Lng,
					},
					OriginalAddress: shipping.Address{
						State:   pickupLocation.State,
						Country: "Philippines",
					},
					Contact: contact,
				},
				DeliveryLocation: shipping.Location{
					Address: location.Address,
					Coordinates: shipping.Coordinates{
						Lat: deliveryCoords.Lat,
						Lng: deliveryCoords.Lng,
					},
					OriginalAddress: shipping.Address{
						State:   location.State,
						Country: "Philippines",
					},
					Contact: contact,
				},
				Package: shipping.Package{
					Weight:      weight,
					Description: "Test package",
				},
				ServiceType: shipping.SERVICE_TYPE_STANDARD,
			}

			quotation, err := ss.GetQuotation(req)

			logs.LogExternalAPICall(ctx, db.GetQueries(), logs.ExternalAPILogParams{
				CheckoutID: nil,
				Service:    "shipping",
				API:        ss.Enum(),
				Endpoint:   "/v3/quotations",
				HTTPMethod: "POST",
				Payload:    req,
				Response:   quotation,
				Error:      err,
			})

			result := QuotationResult{
				Destination: location.Name,
				Weight:      weight + " kg",
				Error:       err,
			}

			if err == nil {
				result.Fee = quotation.Fee
				result.Distance = quotation.DistanceKm
				result.ETA = quotation.EstimatedETA
				if metadata, ok := quotation.Metadata["free_delivery"]; ok {
					result.FreeDelivery = metadata.(bool)
				}
			}

			results = append(results, result)
		}
	}

	printQuotationMatrix(results, weights)

	fmt.Println("\n=== C-Choice Service: Testing Unsupported Operations ===")
}

func testLalamoveService(ctx context.Context) {
	ss := lalamove.MustInit()
	db := database.New(database.DB_MODE_RW)
	geocoder := googlemaps.MustInit(db)

	fmt.Println("=== Lalamove Shipping Service - Matrix Test ===")
	fmt.Println()

	pickupLocation := TestLocation{
		Name:    "General Trias, Cavite",
		Address: "General Trias, Cavite, Philippines",
		State:   "Cavite",
	}

	deliveryLocations := []TestLocation{
		{Name: "Imus, Cavite", Address: "Imus, Cavite, Philippines", State: "Cavite"},
		{Name: "Pasay City", Address: "Pasay City, Metro Manila, Philippines", State: "Metro Manila"},
		{Name: "Quezon City", Address: "Quezon City, Metro Manila, Philippines", State: "Metro Manila"},
	}

	weights := []string{"LESS_THAN_3KG", "LESS_THAN_10KG", "LESS_THAN_20KG"}
	weightLabels := []string{"<3 kg", "<10 kg", "<20 kg"}

	contact := shipping.Contact{
		Name:  "Juan Dela Cruz",
		Phone: "+639171234567",
	}

	pickupCoords, err := geocoder.GeocodeShippingAddress(pickupLocation.Address)
	if err != nil {
		panic(err)
	}

	results := []QuotationResult{}
	for _, location := range deliveryLocations {
		deliveryCoords, err := geocoder.GeocodeShippingAddress(location.Address)
		if err != nil {
			fmt.Printf("Error geocoding %s: %v\n", location.Name, err)
			continue
		}

		for i, weight := range weights {
			req := shipping.ShippingRequest{
				PickupLocation: shipping.Location{
					Address: pickupLocation.Address,
					Coordinates: shipping.Coordinates{
						Lat: pickupCoords.Lat,
						Lng: pickupCoords.Lng,
					},
					Contact: contact,
				},
				DeliveryLocation: shipping.Location{
					Address: location.Address,
					Coordinates: shipping.Coordinates{
						Lat: deliveryCoords.Lat,
						Lng: deliveryCoords.Lng,
					},
					Contact: contact,
				},
				Package: shipping.Package{
					Weight:      weight,
					Description: "Test package",
				},
				ServiceType: shipping.SERVICE_TYPE_MOTORCYCLE,
				ScheduledAt: time.Now().UTC().Add(1 * time.Hour).Format("2006-01-02T15:04:05Z"),
			}

			quotation, err := ss.GetQuotation(req)

			logs.LogExternalAPICall(ctx, db.GetQueries(), logs.ExternalAPILogParams{
				CheckoutID: nil,
				Service:    "shipping",
				API:        ss.Enum(),
				Endpoint:   "/v3/quotations",
				HTTPMethod: "POST",
				Payload:    req,
				Response:   quotation,
				Error:      err,
			})

			result := QuotationResult{
				Destination: location.Name,
				Weight:      weightLabels[i],
				Error:       err,
			}

			if err == nil {
				result.Fee = quotation.Fee
				result.Distance = quotation.DistanceKm
				result.ETA = quotation.EstimatedETA
			}

			results = append(results, result)
		}
	}

	printQuotationMatrix(results, weightLabels)
	fmt.Println("\n=== Test Completed ===")
}

func printQuotationMatrix(results []QuotationResult, weights []string) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                         QUOTATION RESULTS MATRIX                          ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ %-20s │ %-10s │ %-12s │ %-10s │ %-8s ║\n", "Destination", "Weight", "Fee (PHP)", "Distance", "ETA")
	fmt.Println("╠════════════════════════════════════════════════════════════════════════════╣")

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("║ %-20s │ %-10s │ %-12s │ %-10s │ %-8s ║\n",
				result.Destination,
				result.Weight,
				"ERROR",
				"-",
				"-",
			)
			continue
		}

		feeStr := fmt.Sprintf("%.2f", result.Fee)
		if result.FreeDelivery {
			feeStr = "FREE"
		}
		distStr := fmt.Sprintf("%.2f km", result.Distance)
		etaStr := fmt.Sprintf("%d min", result.ETA)

		fmt.Printf("║ %-20s │ %-10s │ %-12s │ %-10s │ %-8s ║\n",
			result.Destination,
			result.Weight,
			feeStr,
			distStr,
			etaStr,
		)
	}

	fmt.Println("╚════════════════════════════════════════════════════════════════════════════╝")
}
