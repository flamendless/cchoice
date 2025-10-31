package orders

import (
	"cchoice/internal/cart"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/geocoding"
	"cchoice/internal/logs"
	"cchoice/internal/payments"
	"cchoice/internal/payments/paymongo"
	"cchoice/internal/requests"
	"cchoice/internal/shipping"
	"cchoice/internal/utils"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

type CreateOrderParams struct {
	CheckoutID              int64
	Checkout                cart.CartCheckout
	CheckoutLines           []queries.GetCheckoutLinesByCheckoutIDRow
	CheckoutSessionResponse payments.CreateCheckoutSessionResponse
	ShippingQuotation       *shipping.ShippingQuotation
	ShippingCoordinates     *shipping.Coordinates
	PaymentGateway          payments.IPaymentGateway
	Geocoder                geocoding.IGeocoder
	Cache                   *fastcache.Cache
	SingleFlight            *singleflight.Group
	Encoder                 encode.IEncode
}

const (
	orderRefPrefix       = "CO-"
	orderRefRandomLength = 4
)

func GenerateUniqueOrderReferenceNumber(ctx context.Context) (string, error) {
	ts := time.Now().UTC().UnixNano()
	tsEnc := strconv.FormatInt(ts, 36)

	randomBytes := make([]byte, orderRefRandomLength)
	if _, err := rand.Read(randomBytes); err != nil {
		tsEnc = strconv.FormatInt(time.Now().UTC().UnixNano(), 36)
		randomHex := strconv.FormatInt(time.Now().UTC().UnixNano(), 16)
		return fmt.Sprintf("%s%s-%s", orderRefPrefix, tsEnc, randomHex[:orderRefRandomLength*2]), nil
	}

	randomHex := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s%s-%s", orderRefPrefix, tsEnc, randomHex), nil
}

func CreateOrderFromCheckout(
	ctx context.Context,
	dbRW database.Service,
	params CreateOrderParams,
) (*queries.TblOrder, string, error) {
	paymongoResponse, ok := params.CheckoutSessionResponse.(*paymongo.CreateCheckoutSessionResponse)
	if !ok {
		return nil, "", errs.ErrPaymentResponse
	}
	checkoutSessionID := paymongoResponse.Data.ID

	totalAmount := int64(0)
	for _, checkoutLine := range params.CheckoutLines {
		totalAmount += checkoutLine.UnitPriceWithVat * checkoutLine.Quantity
	}
	if params.ShippingQuotation != nil {
		totalAmount += int64(params.ShippingQuotation.Fee * 100)
	}

	placeholderPayment := queries.CreateCheckoutPaymentParams{
		ID:                     checkoutSessionID,
		Gateway:                params.PaymentGateway.GatewayEnum().String(),
		CheckoutID:             params.CheckoutID,
		Status:                 "pending",
		Description:            "Order payment - status pending",
		TotalAmount:            totalAmount,
		CheckoutUrl:            paymongoResponse.Data.Attributes.CheckoutURL,
		ClientKey:              paymongoResponse.Data.Attributes.ClientKey,
		ReferenceNumber:        paymongoResponse.Data.Attributes.ReferenceNumber,
		PaymentStatus:          "pending",
		PaymentMethodType:      strings.Join(paymongoResponse.Data.Attributes.PaymentMethodTypes, ","),
		PaidAt:                 time.Time{},
		MetadataRemarks:        "",
		MetadataNotes:          "",
		MetadataCustomerNumber: "",
	}

	checkoutPayment, err := dbRW.GetQueries().CreateCheckoutPayment(ctx, placeholderPayment)
	if err != nil {
		return nil, "", err
	}

	var shippingLat, shippingLng, shippingFormattedAddr, shippingPlaceID sql.NullString
	if params.ShippingCoordinates != nil && params.ShippingCoordinates.Lat != "" && params.ShippingCoordinates.Lng != "" {
		shippingLat = sql.NullString{String: params.ShippingCoordinates.Lat, Valid: true}
		shippingLng = sql.NullString{String: params.ShippingCoordinates.Lng, Valid: true}
	} else {
		shippingAddress := strings.Join([]string{
			params.Checkout.AddressLine1,
			params.Checkout.AddressLine2,
			params.Checkout.Barangay,
			params.Checkout.City,
			params.Checkout.Province,
			params.Checkout.Postal,
			"Philippines",
		}, ", ")
		shippingAddress = strings.ReplaceAll(shippingAddress, ", , ", ", ")

		if coordinates, err := requests.GetGeocodingCoordinates(params.Cache, params.SingleFlight, params.Geocoder, shippingAddress); err == nil {
			shippingLat = sql.NullString{String: coordinates.Lat, Valid: coordinates.Lat != ""}
			shippingLng = sql.NullString{String: coordinates.Lng, Valid: coordinates.Lng != ""}
		}
	}

	subtotal := utils.NewMoney(0, "PHP")
	deliveryFee := utils.NewMoney(0, "PHP")
	totalDiscounts := utils.NewMoney(0, "PHP")

	if params.ShippingQuotation != nil {
		deliveryFee = utils.NewMoney(int64(params.ShippingQuotation.Fee*100), params.ShippingQuotation.Currency)
	}

	for _, checkoutLine := range params.CheckoutLines {
		sub := utils.NewMoney(checkoutLine.UnitPriceWithVat, checkoutLine.UnitPriceWithVatCurrency).Multiply(checkoutLine.Quantity)
		newSubtotal, err := subtotal.Add(sub)
		if err != nil {
			return nil, "", err
		}
		subtotal = newSubtotal
	}

	total, _ := subtotal.Add(deliveryFee)
	total, _ = total.Subtract(totalDiscounts)

	dbCheckoutLineIDs := make(map[int64]bool)
	for _, encodedID := range params.Checkout.CheckoutIDs {
		dbCheckoutLineIDs[params.Encoder.Decode(encodedID)] = true
	}

	orderNumber, err := GenerateUniqueOrderReferenceNumber(ctx)
	if err != nil {
		return nil, "", err
	}

	orderParams := queries.CreateOrderParams{
		CheckoutID:               params.CheckoutID,
		CheckoutPaymentID:        checkoutPayment.ID,
		OrderNumber:              orderNumber,
		Status:                   enums.ORDER_STATUS_PENDING.String(),
		CustomerName:             params.Checkout.Email,
		CustomerEmail:            params.Checkout.Email,
		CustomerPhone:            params.Checkout.MobileNo,
		BillingAddressLine1:      params.Checkout.AddressLine1,
		BillingAddressLine2:      params.Checkout.AddressLine2,
		BillingCity:              params.Checkout.City,
		BillingState:             params.Checkout.Province,
		BillingPostalCode:        params.Checkout.Postal,
		BillingCountry:           "PH",
		BillingLatitude:          shippingLat,
		BillingLongitude:         shippingLng,
		BillingFormattedAddress:  shippingFormattedAddr,
		BillingPlaceID:           shippingPlaceID,
		ShippingAddressLine1:     params.Checkout.AddressLine1,
		ShippingAddressLine2:     params.Checkout.AddressLine2,
		ShippingCity:             params.Checkout.City,
		ShippingState:            params.Checkout.Province,
		ShippingPostalCode:       params.Checkout.Postal,
		ShippingCountry:          "PH",
		ShippingLatitude:         shippingLat,
		ShippingLongitude:        shippingLng,
		ShippingFormattedAddress: shippingFormattedAddr,
		ShippingPlaceID:          shippingPlaceID,
		SubtotalAmount:           subtotal.Amount(),
		ShippingAmount:           deliveryFee.Amount(),
		DiscountAmount:           totalDiscounts.Amount(),
		TotalAmount:              total.Amount(),
		Currency:                 "PHP",
		ShippingService:          sql.NullString{Valid: false},
		ShippingOrderID:          sql.NullString{Valid: false},
		ShippingTrackingNumber:   sql.NullString{Valid: false},
		Notes:                    sql.NullString{Valid: false},
		Remarks:                  sql.NullString{Valid: false},
	}

	order, err := dbRW.GetQueries().CreateOrder(ctx, orderParams)
	if err != nil {
		return nil, "", err
	}

	for _, checkoutLine := range params.CheckoutLines {
		if !dbCheckoutLineIDs[checkoutLine.ID] {
			continue
		}

		unitPrice := utils.NewMoney(checkoutLine.UnitPriceWithVat, checkoutLine.UnitPriceWithVatCurrency)
		totalPrice := unitPrice.Multiply(checkoutLine.Quantity)

		orderLineParams := queries.CreateOrderLineParams{
			OrderID:        order.ID,
			CheckoutLineID: checkoutLine.ID,
			ProductID:      checkoutLine.ProductID,
			Name:           checkoutLine.Name,
			Serial:         "", // TODO: Get serial from product if needed
			Description:    checkoutLine.Description.String,
			UnitPrice:      unitPrice.Amount(),
			Quantity:       checkoutLine.Quantity,
			TotalPrice:     totalPrice.Amount(),
			Currency:       checkoutLine.UnitPriceWithVatCurrency,
		}

		if _, err := dbRW.GetQueries().CreateOrderLine(ctx, orderLineParams); err != nil {
			logs.Log().Error(
				"Failed to create order line",
				zap.Error(err),
				zap.Int64("order_id", order.ID),
				zap.Int64("checkout_line_id", checkoutLine.ID),
			)
		}
	}

	logs.Log().Info(
		"Created order from checkout",
		zap.Int64("order_id", order.ID),
		zap.String("order_number", order.OrderNumber),
		zap.Int64("checkout_id", params.CheckoutID),
	)

	return &order, paymongoResponse.Data.Attributes.CheckoutURL, nil
}
