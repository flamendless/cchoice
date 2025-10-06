package cmd

import (
	"cchoice/internal/geocoding"
	"cchoice/internal/geocoding/googlemaps"
	"cchoice/internal/shipping"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(testGeocodingCmd)
}

var testGeocodingCmd = &cobra.Command{
	Use:   "test_geocoding",
	Short: "Test geocoding functionality",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Testing Google Maps Geocoding Integration")
		fmt.Println("=========================================")

		geocoder := googlemaps.MustInit()

		testAddresses := []string{
			"EDSA cor. Ortigas Avenue, Pasig City, Metro Manila, Philippines",
			"Ayala Avenue, Makati City, Metro Manila, Philippines",
			"SM Mall of Asia, Pasay City, Metro Manila, Philippines",
			"Bonifacio Global City, Taguig City, Metro Manila, Philippines",
			"University of the Philippines Diliman, Quezon City, Metro Manila, Philippines",
		}

		fmt.Printf("\n1. Testing Direct Geocoding\n")
		fmt.Println("---------------------------")

		for i, address := range testAddresses {
			fmt.Printf("\nTest %d: %s\n", i+1, address)

			req := geocoding.GeocodeRequest{
				Address: address,
				ComponentFilter: map[string]string{
					"country": "PH",
				},
			}

			result, err := geocoder.Geocode(req)
			if err != nil {
				panic(err)
			}

			fmt.Printf("  Coordinates: %s, %s\n", result.Coordinates.Lat, result.Coordinates.Lng)
			fmt.Printf("  Formatted Address: %s\n", result.FormattedAddress)
			if result.PlaceID != "" {
				fmt.Printf("  Place ID: %s\n", result.PlaceID)
			}
		}

		fmt.Printf("\n2. Testing Shipping Address Geocoding\n")
		fmt.Println("--------------------------------------")

		for i, address := range testAddresses[:2] {
			fmt.Printf("\nShipping Test %d: %s\n", i+1, address)

			coordinates, err := geocoder.GeocodeShippingAddress(address)
			if err != nil {
				panic(err)
			}

			fmt.Printf("  Lat: %s, Lng: %s\n", coordinates.Lat, coordinates.Lng)
		}

		fmt.Printf("\n3. Testing Reverse Geocoding\n")
		fmt.Println("----------------------------")

		testCoords := shipping.Coordinates{
			Lat: "14.5832",
			Lng: "120.9794",
		}

		fmt.Printf("\nReverse geocoding coordinates: %s, %s\n", testCoords.Lat, testCoords.Lng)

		reverseReq := geocoding.ReverseGeocodeRequest{
			Coordinates: geocoding.Coordinates{
				Lat: testCoords.Lat,
				Lng: testCoords.Lng,
			},
			Language: "en",
		}

		reverseResult, err := geocoder.ReverseGeocode(reverseReq)
		if err != nil {
			panic(err)
		}
		fmt.Printf("  Address: %s\n", reverseResult.FormattedAddress)

		fmt.Printf("\n4. Testing Proper Geocoding Integration with Shipping Service\n")
		fmt.Println("------------------------------------------------------------")

		fmt.Printf("\nStep 1: Create shipping request with addresses (no coordinates)\n")
		shippingReq := shipping.ShippingRequest{
			PickupLocation: shipping.Location{
				Address: "EDSA cor. Ortigas Avenue, Pasig City, Metro Manila, Philippines",
				Contact: shipping.Contact{
					Name:  "John Doe",
					Phone: "+639123456789",
				},
			},
			DeliveryLocation: shipping.Location{
				Address: "Ayala Avenue, Makati City, Metro Manila, Philippines",
				Contact: shipping.Contact{
					Name:  "Jane Smith",
					Phone: "+639987654321",
				},
			},
			Package: shipping.Package{
				Weight:      "2.5",
				Description: "Test package",
			},
			ServiceType: shipping.SERVICE_TYPE_MOTORCYCLE,
		}

		fmt.Printf("  Pickup Address: %s\n", shippingReq.PickupLocation.Address)
		fmt.Printf("  Delivery Address: %s\n", shippingReq.DeliveryLocation.Address)
		fmt.Printf("  Initial Coordinates: Empty\n")

		fmt.Printf("\nStep 2: Use geocoding to get coordinates for both locations\n")
		pickupCoords, err := geocoder.GeocodeShippingAddress(shippingReq.PickupLocation.Address)
		if err != nil {
			panic(err)
		}
		shippingReq.PickupLocation.Coordinates = shipping.Coordinates{
			Lat: pickupCoords.Lat,
			Lng: pickupCoords.Lng,
		}
		fmt.Printf("  Pickup Coordinates: %s, %s\n", pickupCoords.Lat, pickupCoords.Lng)

		deliveryCoords, err := geocoder.GeocodeShippingAddress(shippingReq.DeliveryLocation.Address)
		if err != nil {
			panic(err)
		}
		shippingReq.DeliveryLocation.Coordinates = shipping.Coordinates{
			Lat: deliveryCoords.Lat,
			Lng: deliveryCoords.Lng,
		}
		fmt.Printf("  Delivery Coordinates: %s, %s\n", deliveryCoords.Lat, deliveryCoords.Lng)

		fmt.Printf("\nStep 3: Pass prepared request to shipping service\n")
		fmt.Printf("  ✓ Request is now ready for shipping service (has coordinates)\n")

		reqJSON, _ := json.MarshalIndent(shippingReq, "", "  ")
		fmt.Printf("\nComplete Shipping Request:\n%s\n", reqJSON)

		fmt.Printf("\nStep 4: Demonstrate with actual shipping service call (Lalamove)\n")
		fmt.Printf("  Note: This is the correct approach - geocoding happens at the application level\n")
		fmt.Printf("  The shipping service receives a request that already has coordinates\n")

		fmt.Printf("\n5. Testing Helper Utility Functions\n")
		fmt.Println("------------------------------------")

		helper := shipping.NewGeocodingHelper(geocoder)

		fmt.Printf("\nTesting EnsureCoordinates helper:\n")
		testReq := shipping.ShippingRequest{
			PickupLocation: shipping.Location{
				Address: "SM Mall of Asia, Pasay City, Metro Manila, Philippines",
				Contact: shipping.Contact{Name: "Test User", Phone: "+639123456789"},
			},
			DeliveryLocation: shipping.Location{
				Address: "Bonifacio Global City, Taguig City, Metro Manila, Philippines",
				Contact: shipping.Contact{Name: "Test Recipient", Phone: "+639987654321"},
			},
			Package: shipping.Package{Weight: "1.0", Description: "Test"},
		}

		err = helper.EnsureCoordinates(&testReq)
		if err != nil {
			panic(err)
		}
		fmt.Printf("  ✓ Coordinates ensured successfully\n")
		fmt.Printf("  Pickup: %s, %s\n", testReq.PickupLocation.Coordinates.Lat, testReq.PickupLocation.Coordinates.Lng)
		fmt.Printf("  Delivery: %s, %s\n", testReq.DeliveryLocation.Coordinates.Lat, testReq.DeliveryLocation.Coordinates.Lng)

		fmt.Printf("\nTesting request validation:\n")
		err = helper.ValidateShippingRequest(&testReq)
		if err != nil {
			panic(err)
		}
		fmt.Printf("  ✓ Shipping request is valid\n")
		fmt.Printf("\nGeocoding test completed!\n")
	},
}
