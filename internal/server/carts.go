package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"

	"cchoice/cmd/web/components"
	"cchoice/cmd/web/models"
	"cchoice/internal/cart"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/images"
	"cchoice/internal/logs"
	"cchoice/internal/orders"
	"cchoice/internal/payments"
	"cchoice/internal/shipping"
	"cchoice/internal/storage"
	"cchoice/internal/utils"

	"github.com/Rhymond/go-money"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func AddCartsHandlers(s *Server, r chi.Router) {
	r.Get("/carts", s.cartsPageHandler)
	r.Get("/carts/summary", s.getCartSummaryHandler)
	r.Get("/carts/lines", s.cartLinesHandler)
	r.Get("/carts/lines/count", s.getCartLinesCountHandler)
	r.Post("/carts/lines", s.addProductToCartHandler)
	r.Delete("/carts/lines/{checkoutline_id}", s.removeProductFromCartHandler)
	r.Patch("/carts/lines/{checkoutline_id}", s.updateCartLinesQtyHandler)
	r.Patch("/carts/lines/{checkoutline_id}/toggle", s.toggleCartLineCheckboxHandler)
	r.Get("/carts/payment-methods", s.cartsPaymentMethodsHandler)
	r.Post("/carts/finalize", s.cartsFinalizeHandler)
}

type cartSummaryData struct {
	Subtotal      string
	TotalDiscount string
	TotalWeight   string
	DeliveryFee   string
	Total         string
	DeliveryETA   string
}

func (s *Server) calculateCartSummary(ctx context.Context) (cartSummaryData, error) {
	token := s.sessionManager.Token(ctx)
	checkedItems := GetCheckedItems(ctx, s.sessionManager)

	var checkoutLines []queries.GetCheckoutLinesByCheckoutIDRow
	var err error

	if len(checkedItems) > 0 {
		checkoutLines, err = cart.GetCheckedCheckoutLines(ctx, s.dbRO, token, checkedItems, s.encoder)
	} else {
		checkoutLines = []queries.GetCheckoutLinesByCheckoutIDRow{}
	}

	if err != nil {
		return cartSummaryData{}, err
	}

	var subtotal = utils.NewMoney(0, "PHP")
	deliveryFee := utils.NewMoney(0, "PHP")
	totalDiscounts := utils.NewMoney(0, "PHP")

	if quotation, ok := s.sessionManager.Get(ctx, skShippingQuotation).(*shipping.ShippingQuotation); ok && quotation != nil {
		deliveryFee = utils.NewMoney(int64(quotation.Fee*100), quotation.Currency)
	}

	for _, checkoutLine := range checkoutLines {
		_, discountedPrice, _ := utils.GetOrigAndDiscounted(
			checkoutLine.IsOnSale,
			checkoutLine.UnitPriceWithVat,
			checkoutLine.UnitPriceWithVatCurrency,
			checkoutLine.SalePriceWithVat,
			checkoutLine.SalePriceWithVatCurrency,
		)

		sub := discountedPrice.Multiply(checkoutLine.Quantity)
		newSubtotal, err := subtotal.Add(sub)
		if err != nil {
			continue
		}
		subtotal = newSubtotal
	}

	total, _ := subtotal.Add(deliveryFee)
	total, _ = total.Subtract(totalDiscounts)
	totalWeightKg, _ := utils.CalculateTotalWeightFromCheckoutLines(checkoutLines)

	deliveryETA := ""
	if shippingReq, ok := s.sessionManager.Get(ctx, skShippingRequest).(*shipping.ShippingRequest); ok && shippingReq != nil {
		province := shippingReq.DeliveryLocation.OriginalAddress.State
		deliveryETA = s.shippingService.GetDeliveryETA(ctx, province)
	}

	return cartSummaryData{
		Subtotal:      subtotal.Display(),
		TotalDiscount: "- " + totalDiscounts.Display(),
		TotalWeight:   totalWeightKg + " kg",
		DeliveryFee:   deliveryFee.Display(),
		Total:         total.Display(),
		DeliveryETA:   deliveryETA,
	}, nil
}

func (s *Server) generateCartSummaryComponent(ctx context.Context) templ.Component {
	summaryData, err := s.calculateCartSummary(ctx)
	if err != nil {
		return components.CartSummaryContentEmpty()
	}

	return components.CartSummaryContent(
		summaryData.Subtotal,
		summaryData.TotalDiscount,
		summaryData.TotalWeight,
		summaryData.DeliveryFee,
		summaryData.Total,
		summaryData.DeliveryETA,
	)
}

func (s *Server) cartsPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Cart Page Handler]"
	ctx := r.Context()

	checkoutlineProductIDs, ok := s.sessionManager.Get(ctx, skCheckoutLineProductIDs).([]string)
	if len(checkoutlineProductIDs) == 0 {
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	token := s.sessionManager.Token(ctx)
	if !ok {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("token", token),
			zap.Error(errs.ErrSessionCheckoutLineProductIDs),
		)
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	checkoutID, err := cart.CreateCart(
		ctx,
		s.dbRW.GetQueries(),
		s.encoder,
		token,
		checkoutlineProductIDs,
	)
	if err != nil || checkoutID == -1 {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("token", token),
			zap.Error(err),
		)
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	summaryContent := s.generateCartSummaryComponent(ctx)
	if err := components.CartPage(components.CartPageBody(summaryContent)).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) cartLinesHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Cart Lines Handler]"
	ctx := r.Context()
	token := s.sessionManager.Token(ctx)

	checkoutLines, err := cart.GetCheckoutLines(ctx, s.dbRO, token)
	if err != nil {
		logs.LogCtx(ctx).Warn(
			logtag,
			zap.String("token", token),
			zap.Error(err),
			zap.Error(errs.ErrCartMissingCheckoutLines),
		)
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if !s.sessionManager.Exists(ctx, skCheckedItems) {
		var checkedItems []string
		for _, checkoutLine := range checkoutLines {
			checkedItems = append(checkedItems, s.encoder.Encode(checkoutLine.ID))
		}
		SetCheckedItems(ctx, s.sessionManager, checkedItems)
	}
	checkedItems := GetCheckedItems(ctx, s.sessionManager)

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	type checkoutLineWithImage struct {
		line    queries.GetCheckoutLinesByCheckoutIDRow
		imgData string
		index   int
	}

	var mu sync.Mutex
	lineResults := make([]checkoutLineWithImage, 0, len(checkoutLines))

	for i := range checkoutLines {
		g.Go(func() error {
			var imgData string
			if !strings.HasSuffix(checkoutLines[i].ThumbnailPath, constants.EmptyImageFilename) {
				if imgDataB64, err := images.GetImageDataB64(s.cache, s.productImageFS, checkoutLines[i].ThumbnailPath, images.IMAGE_FORMAT_WEBP); err == nil {
					imgData = imgDataB64
				} else {
					logs.LogCtx(gctx).Error(
						logtag,
						zap.Error(err),
					)
				}
			}

			mu.Lock()
			lineResults = append(lineResults, checkoutLineWithImage{
				line:    checkoutLines[i],
				imgData: imgData,
				index:   i,
			})
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slices.SortFunc(lineResults, func(a, b checkoutLineWithImage) int {
		return a.index - b.index
	})

	for _, result := range lineResults {
		checkoutLine := result.line

		origPrice, discountedPrice, discountPercentage := utils.GetOrigAndDiscounted(
			checkoutLine.IsOnSale,
			checkoutLine.UnitPriceWithVat,
			checkoutLine.UnitPriceWithVatCurrency,
			checkoutLine.SalePriceWithVat,
			checkoutLine.SalePriceWithVatCurrency,
		)

		encodedID := s.encoder.Encode(checkoutLine.ID)
		isChecked := slices.Contains(checkedItems, encodedID)

		cl := models.CheckoutLine{
			ID:                 encodedID,
			CheckoutID:         s.encoder.Encode(checkoutLine.CheckoutID),
			ProductID:          s.encoder.Encode(checkoutLine.ProductID),
			Name:               checkoutLine.Name,
			BrandName:          checkoutLine.BrandName,
			Quantity:           checkoutLine.Quantity,
			ThumbnailPath:      checkoutLine.ThumbnailPath,
			CDNURL:             s.GetCDNURL(checkoutLine.ThumbnailPath),
			CDNURL1280:         s.GetCDNURL(constants.ToPath1280(checkoutLine.ThumbnailPath)),
			OrigPrice:          *origPrice,
			Price:              *discountedPrice,
			Total:              *discountedPrice.Multiply(checkoutLine.Quantity),
			Checked:            isChecked,
			DiscountPercentage: discountPercentage,
		}

		if weightKg, err := utils.ConvertWeightToKg(checkoutLine.Weight, checkoutLine.WeightUnit); err == nil {
			cl.WeightKg = weightKg
			cl.WeightDisplay = fmt.Sprintf("%.2f kg", weightKg)
		}

		if err := components.CartCheckoutLineItem(cl).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			continue
		}
	}
}

func (s *Server) addProductToCartHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Add Product To Cart Handler]"
	ctx := r.Context()
	token := s.sessionManager.Token(ctx)

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	productID := r.Form.Get("product_id")
	if productID == "" {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(errs.ErrInvalidParams),
		)
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	dbProductID := s.encoder.Decode(productID)
	if _, err := s.dbRO.GetQueries().CheckProductExistsByID(ctx, dbProductID); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("token", token),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	checkoutLineProductIDs, err := AddToCheckoutLineProductIDs(ctx, s.sessionManager, productID)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("token", token),
			zap.String("product id", productID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("token", token),
		zap.Strings("checkout line product ids", checkoutLineProductIDs),
	)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) removeProductFromCartHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Remove Product From Cart Handler]"
	ctx := r.Context()
	token := s.sessionManager.Token(ctx)

	checkoutLineID := chi.URLParam(r, "checkoutline_id")
	dbCheckoutLineID := s.encoder.Decode(checkoutLineID)

	checkoutLine, err := s.dbRO.GetQueries().GetCheckoutLineByID(ctx, dbCheckoutLineID)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("token", token),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := RemoveFromCheckoutLineProductIDs(
		ctx,
		s.sessionManager,
		s.encoder.Encode(checkoutLine.ProductID),
	); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("token", token),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.dbRW.GetQueries().DeleteCheckoutLineByID(ctx, dbCheckoutLineID); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("token", token),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	remaining, err := s.dbRO.GetQueries().CountCheckoutLineByCheckoutID(ctx, checkoutLine.CheckoutID)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("token", token),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if remaining == 0 {
		w.Header().Set("HX-Redirect", utils.URL("/carts"))
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getCartSummaryHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Get Cart Summary Handler]"
	ctx := r.Context()

	data := r.URL.Query().Get("data")
	switch data {
	default:
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("data", data),
			zap.Error(errs.ErrInvalidParams),
		)
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)

	case "summary_total":
		summaryData, err := s.calculateCartSummary(ctx)
		if err != nil {
			token := s.sessionManager.Token(ctx)
			logs.LogCtx(ctx).Warn(
				logtag,
				zap.String("token", token),
				zap.Error(err),
				zap.Error(errs.ErrCartMissingCheckoutLines),
			)
			if _, err := w.Write([]byte("0 Items")); err != nil {
				logs.LogCtx(ctx).Error(
					logtag,
					zap.Error(err),
				)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		var renderErrs error
		renderErrs = errors.Join(renderErrs, components.CartSummaryRow("Subtotal", summaryData.Subtotal, "text-gray-500").Render(ctx, w))
		renderErrs = errors.Join(renderErrs, components.CartSummaryRow("Total Discount", summaryData.TotalDiscount, "text-red-500").Render(ctx, w))
		renderErrs = errors.Join(renderErrs, components.CartSummaryRow("Total Weight", summaryData.TotalWeight, "text-gray-500").Render(ctx, w))
		renderErrs = errors.Join(renderErrs, components.CartSummaryRowWithID("delivery-fee-row", "Delivery Fee", summaryData.DeliveryFee, "text-gray-500").Render(ctx, w))
		renderErrs = errors.Join(renderErrs, components.CartSummaryRowWithID("delivery-eta-row", "Estimated Delivery Time", summaryData.DeliveryETA, "text-gray-500").Render(ctx, w))
		renderErrs = errors.Join(renderErrs, components.HR().Render(ctx, w))
		renderErrs = errors.Join(renderErrs, components.CartSummaryRow("Total", summaryData.Total).Render(ctx, w))

		if renderErrs != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(renderErrs),
			)
			http.Error(w, renderErrs.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) getCartLinesCountHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Get Cart Lines Count Handler]"
	ctx := r.Context()

	set := map[string]bool{}
	if productIDs, ok := s.sessionManager.Get(ctx, skCheckoutLineProductIDs).([]string); ok {
		for _, productID := range productIDs {
			if _, exists := set[productID]; !exists {
				set[productID] = true
			}
		}
	}
	if _, err := w.Write(fmt.Appendf(nil, "%d", len(set))); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) updateCartLinesQtyHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Update Cart Lines Qty Handler]"
	ctx := r.Context()

	checkoutLineID := chi.URLParam(r, "checkoutline_id")
	dbCheckoutLineID := s.encoder.Decode(checkoutLineID)

	qty := 0
	if r.URL.Query().Get("dec") == "1" {
		qty = -1
	} else if r.URL.Query().Get("inc") == "1" {
		qty = 1
	}

	if qty == 0 {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(errs.ErrInvalidParams),
		)
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	newQty, err := s.dbRW.GetQueries().UpdateCheckoutLineQtyByID(
		ctx,
		queries.UpdateCheckoutLineQtyByIDParams{
			ID:       dbCheckoutLineID,
			Quantity: int64(qty),
		},
	)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(fmt.Appendf(nil, "Qty: %d", newQty)); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) toggleCartLineCheckboxHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Toggle Cart Line Checkbox Handler]"
	ctx := r.Context()

	checkoutLineID := chi.URLParam(r, "checkoutline_id")
	token := s.sessionManager.Token(ctx)
	checkedItems := ToggleCheckedItem(ctx, s.sessionManager, checkoutLineID)

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("token", token),
		zap.String("checkoutline id", checkoutLineID),
		zap.Strings("checked items", checkedItems),
	)

	w.WriteHeader(http.StatusOK)
}

func (s *Server) cartsFinalizeHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Cart Finalize Handler]"
	ctx := r.Context()

	var cartCheckout cart.CartCheckout
	if err := utils.FormToStruct(r, &cartCheckout); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := s.sessionManager.Token(ctx)
	if err := cart.KeepItemsInCheckoutLines(ctx, s.dbRW, token, cartCheckout.ToDBCheckoutIDs(s.encoder)); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("token", token),
			zap.Any("form", cartCheckout),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	checkoutLines, err := cart.GetCheckoutLines(ctx, s.dbRO, token)
	if err != nil || len(checkoutLines) == 0 {
		logs.LogCtx(ctx).Warn(
			logtag,
			zap.String("token", token),
			zap.Error(err),
			zap.Error(errs.ErrCartMissingCheckoutLines),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	checkoutID, err := s.dbRO.GetQueries().GetCheckoutIDBySessionID(ctx, token)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("token", token),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch s.paymentGateway.GatewayEnum() {
	case payments.PAYMENT_GATEWAY_PAYMONGO:
		paymentMethods := []payments.PaymentMethod{
			payments.ParsePaymentMethodToEnum(cartCheckout.PaymentMethod),
		}
		billing := payments.Billing{
			Address: payments.Address{
				Line1:      cartCheckout.AddressLine1,
				Line2:      cartCheckout.AddressLine2,
				City:       cartCheckout.City,
				State:      cartCheckout.Province,
				PostalCode: cartCheckout.Postal,
				Country:    "PH",
			},
			Name:  cartCheckout.FullName,
			Email: cartCheckout.Email,
			Phone: cartCheckout.MobileNo,
		}
		lineItems := make([]payments.LineItem, 0, len(cartCheckout.CheckoutIDs))
		for _, checkoutLine := range checkoutLines {
			imageURL, err := s.GetProductImageProxyURL(ctx, checkoutLine.ThumbnailPath, "256x256")
			if err != nil {
				logs.LogCtx(ctx).Warn(
					logtag,
					zap.String("thumbnail_path", checkoutLine.ThumbnailPath),
					zap.Error(err),
				)
			}

			_, discountedPrice, _ := utils.GetOrigAndDiscounted(
				checkoutLine.IsOnSale,
				checkoutLine.UnitPriceWithVat,
				checkoutLine.UnitPriceWithVatCurrency,
				checkoutLine.SalePriceWithVat,
				checkoutLine.SalePriceWithVatCurrency,
			)

			lineItems = append(lineItems, payments.LineItem{
				Amount:      int32(discountedPrice.Amount()),
				Currency:    money.PHP,
				Description: checkoutLine.Description.String,
				Images:      []string{imageURL},
				Name:        checkoutLine.Name,
				Quantity:    int32(checkoutLine.Quantity),
			})
		}

		var shippingQuotation *shipping.ShippingQuotation
		if quotation, ok := s.sessionManager.Get(ctx, skShippingQuotation).(*shipping.ShippingQuotation); ok && quotation != nil {
			shippingQuotation = quotation
		}

		var shippingCoordinates *shipping.Coordinates
		var deliveryETA string
		if shippingReq, ok := s.sessionManager.Get(ctx, skShippingRequest).(*shipping.ShippingRequest); ok && shippingReq != nil {
			shippingCoordinates = &shippingReq.DeliveryLocation.Coordinates
			deliveryETA = s.shippingService.GetDeliveryETA(ctx, shippingReq.DeliveryLocation.OriginalAddress.State)
		}

		if shippingQuotation.Fee != 0 {
			lineItems = append(lineItems, payments.LineItem{
				Amount:      int32(shippingQuotation.Fee*100),
				Currency:    money.PHP,
				Description: "Shipping Fee",
				// Images:      []string{imageURL}, //TODO: (Brandon)
				Name:        "Shipping Fee",
				Quantity:    1,
			})
		}

		payload := s.paymentGateway.CreatePayload(billing, lineItems, paymentMethods)
		resCheckout, err := s.paymentGateway.CreateCheckoutPaymentSession(payload)
		logs.LogExternalAPICall(ctx, s.dbRW.GetQueries(), logs.ExternalAPILogParams{
			CheckoutID: &checkoutID,
			Service:    "payment",
			API:        s.paymentGateway.GatewayEnum(),
			Endpoint:   "/checkout_sessions",
			HTTPMethod: "POST",
			Payload:    payload,
			Response:   resCheckout,
			Error:      err,
		})
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Any("payload", payload),
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		orderParams := orders.CreateOrderParams{
			CheckoutID:              checkoutID,
			Checkout:                cartCheckout,
			CheckoutLines:           checkoutLines,
			CheckoutSessionResponse: resCheckout,
			ShippingQuotation:       shippingQuotation,
			ShippingCoordinates:     shippingCoordinates,
			DeliveryETA:             deliveryETA,
			PaymentGateway:          s.paymentGateway,
			Geocoder:                s.geocoder,
			Cache:                   s.cache,
			SingleFlight:            &s.SF,
			Encoder:                 s.encoder,
		}

		order, checkoutURL, err := orders.CreateOrderFromCheckout(ctx, s.dbRW, orderParams)
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logs.LogCtx(ctx).Info(
			logtag,
			zap.String("token", token),
			zap.Int64("order_id", order.ID),
			zap.String("order_number", order.OrderNumber),
		)

		// Redirect to payment gateway
		w.Header().Set("HX-Redirect", checkoutURL)
	default:
		err := fmt.Errorf("%s. %w", logtag, errs.ErrServerUnimplementedGateway)
		logs.LogCtx(ctx).Error(
			err.Error(),
			zap.String("gateway", s.paymentGateway.GatewayEnum().String()),
		)
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}
}

func (s *Server) getPaymentImageURL(pm payments.PaymentMethod) string {
	imgPath := pm.GetImagePath()
	if imgPath == "" {
		return ""
	}
	if s.objectStorage != nil && s.objectStorage.ProviderEnum() == storage.STORAGE_PROVIDER_CLOUDFLARE_IMAGES {
		return s.objectStorage.GetPublicURL(imgPath)
	}
	return utils.URL("/assets/image?filename=payments/" + strings.TrimPrefix(imgPath, constants.PathPaymentImages))
}

func (s *Server) cartsPaymentMethodsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Cart Payment Methods Handler]"
	ctx := r.Context()

	cod, err := s.dbRO.GetQueries().GetSettingsCOD(ctx)
	paymentMethods := []models.AvailablePaymentMethod{
		{
			Value:    payments.PAYMENT_METHOD_COD,
			Enabled:  (err == nil && cod),
			ImageURL: s.getPaymentImageURL(payments.PAYMENT_METHOD_COD),
		},
	}

	switch s.paymentGateway.GatewayEnum() {
	case payments.PAYMENT_GATEWAY_PAYMONGO:
		availablePaymongoMethods, err := s.paymentGateway.GetAvailablePaymentMethods()
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.String("gateway", s.paymentGateway.GatewayEnum().String()),
				zap.Error(err),
			)
			var dnsErr *net.DNSError
			if errors.As(err, &dnsErr) && dnsErr.Err == "no such host" {
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		availablePaymentMethods := availablePaymongoMethods.ToPaymentMethods()
		paymongoMethods := s.paymentGateway.GatewayEnum().GetAllPaymentMethods()
		prioritizedPaymentMethods := s.paymentGateway.GatewayEnum().GetPrioritizedPaymentMethods()
		for _, pm := range paymongoMethods {
			enabled := slices.Contains(availablePaymentMethods, pm)
			if !enabled && !slices.Contains(prioritizedPaymentMethods, pm) {
				continue
			}
			paymentMethods = append(paymentMethods, models.AvailablePaymentMethod{
				Value:    pm,
				Enabled:  enabled,
				ImageURL: s.getPaymentImageURL(pm),
			})
		}

	default:
		err := errs.ErrServerUnimplementedGateway
		logs.LogCtx(ctx).Error(
			err.Error(),
			zap.String("gateway", s.paymentGateway.GatewayEnum().String()),
		)
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	// INFO: (Brandon) - Sort payment methods:
	// 1. Enabled methods first (alphabetically)
	// 2. Disabled methods (COD first, then alphabetically)
	slices.SortFunc(paymentMethods, func(a, b models.AvailablePaymentMethod) int {
		if a.Enabled != b.Enabled {
			if a.Enabled {
				return -1
			}
			return 1
		}

		if !a.Enabled && !b.Enabled {
			aCOD := a.Value == payments.PAYMENT_METHOD_COD
			bCOD := b.Value == payments.PAYMENT_METHOD_COD
			if aCOD != bCOD {
				if aCOD {
					return -1
				}
				return 1
			}
		}

		return strings.Compare(a.Value.String(), b.Value.String())
	})

	for _, pm := range paymentMethods {
		if err := components.CartPaymentMethod(pm).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
