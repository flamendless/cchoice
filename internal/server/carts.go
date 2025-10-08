package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"

	"cchoice/cmd/web/components"
	"cchoice/cmd/web/models"
	"cchoice/internal/cart"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/images"
	"cchoice/internal/logs"
	"cchoice/internal/payments"
	"cchoice/internal/shipping"
	"cchoice/internal/utils"

	"github.com/Rhymond/go-money"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddCartsHandlers(s *Server, r chi.Router) {
	r.Get("/carts", s.cartsPageHandler)
	r.Get("/carts/summary", s.getCartSummaryHandler)
	r.Get("/carts/lines", s.cartLinesHandler)
	r.Get("/carts/lines/count", s.getCartLinesCountHandler)
	r.Post("/carts/lines", s.addProductToCartHandler)
	r.Delete("/carts/lines/{checkoutline_id}", s.removeProductFromCartHandler)
	r.Patch("/carts/lines/{checkoutline_id}", s.updateCartLinesQtyHandler)
	r.Get("/carts/payment-methods", s.cartsPaymentMethodsHandler)
	r.Post("/carts/finalize", s.cartsFinalizeHandler)
}

func (s *Server) cartsPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Cart Page Handler]"
	checkoutlineProductIDs, ok := s.sessionManager.Get(r.Context(), skCheckoutLineProductIDs).([]string)
	if len(checkoutlineProductIDs) == 0 {
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	token := s.sessionManager.Token(r.Context())
	if !ok {
		logs.Log().Error(
			logtag,
			zap.Error(errs.ErrSessionCheckoutLineProductIDs),
			zap.String("token", token),
		)
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	checkoutID, err := cart.CreateCart(
		r.Context(),
		s.dbRW.GetQueries(),
		s.encoder,
		token,
		checkoutlineProductIDs,
	)
	if err != nil || checkoutID == -1 {
		logs.Log().Error(logtag, zap.Error(err), zap.String("token", token))
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := components.CartPage(components.CartPageBody()).Render(r.Context(), w); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) cartLinesHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Cart Lines Handler]"
	token := s.sessionManager.Token(r.Context())
	checkoutLines, err := cart.GetCheckoutLines(r.Context(), s.dbRO, token)
	if err != nil {
		logs.Log().Warn(logtag, zap.Error(err), zap.Error(errs.ErrCartMissingCheckoutLines), zap.String("token", token))
		if err := components.CartPage(components.CartPageBodyEmpty()).Render(r.Context(), w); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	for _, checkoutLine := range checkoutLines {
		var imgData string
		if !strings.HasSuffix(checkoutLine.ThumbnailPath, constants.EmptyImageFilename) {
			if imgDataB64, err := images.GetImageDataB64(s.cache, s.fs, checkoutLine.ThumbnailPath, images.IMAGE_FORMAT_WEBP); err == nil {
				imgData = imgDataB64
			}
		}

		price := utils.NewMoney(checkoutLine.UnitPriceWithVat, checkoutLine.UnitPriceWithVatCurrency)

		//TODO: (Brandon) - Discounts/sales
		discountedPrice := utils.NewMoney(checkoutLine.UnitPriceWithVat, checkoutLine.UnitPriceWithVatCurrency)

		cl := models.CheckoutLine{
			ID:              s.encoder.Encode(checkoutLine.ID),
			CheckoutID:      s.encoder.Encode(checkoutLine.CheckoutID),
			ProductID:       s.encoder.Encode(checkoutLine.ProductID),
			Name:            checkoutLine.Name,
			BrandName:       checkoutLine.BrandName,
			Quantity:        checkoutLine.Quantity,
			ThumbnailPath:   checkoutLine.ThumbnailPath,
			ThumbnailData:   imgData,
			Price:           *price,
			DiscountedPrice: *discountedPrice,
			Total:           *discountedPrice.Multiply(checkoutLine.Quantity),
		}

		if err := components.CartCheckoutLineItem(cl).Render(r.Context(), w); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			continue
		}
	}
}

func (s *Server) addProductToCartHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Add Product To Cart Handler]"
	if err := r.ParseForm(); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	productID := r.Form.Get("product_id")
	if productID == "" {
		logs.Log().Error(logtag, zap.Error(errs.ErrInvalidParams))
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	dbProductID := s.encoder.Decode(productID)
	if _, err := s.dbRO.GetQueries().CheckProductExistsByID(r.Context(), dbProductID); err != nil {
		logs.Log().Error(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	checkoutLineProductIDs, err := AddToCheckoutLineProductIDs(r.Context(), s.sessionManager, productID)
	if err != nil {
		logs.Log().Error(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.String("product id", productID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logs.Log().Info(
		logtag,
		zap.String("token", s.sessionManager.Token(r.Context())),
		zap.Strings("checkout line product ids", checkoutLineProductIDs),
	)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) removeProductFromCartHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Remove Product From Cart Handler]"
	checkoutLineID := chi.URLParam(r, "checkoutline_id")
	dbCheckoutLineID := s.encoder.Decode(checkoutLineID)

	checkoutLine, err := s.dbRO.GetQueries().GetCheckoutLineByID(r.Context(), dbCheckoutLineID)
	if err != nil {
		logs.Log().Error(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := RemoveFromCheckoutLineProductIDs(
		r.Context(),
		s.sessionManager,
		s.encoder.Encode(checkoutLine.ProductID),
	); err != nil {
		logs.Log().Error(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.dbRW.GetQueries().DeleteCheckoutLineByID(r.Context(), dbCheckoutLineID); err != nil {
		logs.Log().Error(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	remaining, err := s.dbRO.GetQueries().CountCheckoutLineByCheckoutID(r.Context(), checkoutLine.CheckoutID)
	if err != nil {
		logs.Log().Error(
			logtag,
			zap.String("token", s.sessionManager.Token(r.Context())),
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if remaining == 0 {
		w.Header().Set("HX-Redirect", "/cchoice/carts")
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getCartSummaryHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Get Cart Summary Handler]"
	data := r.URL.Query().Get("data")
	switch data {
	default:
		logs.Log().Error(
			logtag,
			zap.Error(errs.ErrInvalidParams),
			zap.String("data", data),
		)
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)

	case "bar_items":
		token := s.sessionManager.Token(r.Context())
		checkoutLines, err := cart.GetCheckoutLines(r.Context(), s.dbRO, token)
		if err != nil {
			logs.Log().Warn(logtag, zap.Error(err), zap.Error(errs.ErrCartMissingCheckoutLines), zap.String("token", token))
			if _, err := w.Write([]byte("0 Items")); err != nil {
				logs.Log().Error(logtag, zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		count := 0
		for _, checkoutLine := range checkoutLines {
			count += int(checkoutLine.Quantity)
		}
		if _, err := w.Write(fmt.Appendf(nil, "%d Items", count)); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case "summary_total":
		token := s.sessionManager.Token(r.Context())
		checkoutLines, err := cart.GetCheckoutLines(r.Context(), s.dbRO, token)
		if err != nil {
			logs.Log().Warn(logtag, zap.Error(err), zap.Error(errs.ErrCartMissingCheckoutLines), zap.String("token", token))
			if _, err := w.Write([]byte("0 Items")); err != nil {
				logs.Log().Error(logtag, zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		var errs error
		subtotal := utils.NewMoney(0, "PHP")
		deliveryFee := utils.NewMoney(0, "PHP")
		totalDiscounts := utils.NewMoney(0, "PHP")

		if quotation, ok := s.sessionManager.Get(r.Context(), skShippingQuotation).(*shipping.ShippingQuotation); ok && quotation != nil {
			deliveryFee = utils.NewMoney(int64(quotation.Fee*100), quotation.Currency)
		}

		for _, checkoutLine := range checkoutLines {
			sub := utils.NewMoney(checkoutLine.UnitPriceWithVat, checkoutLine.UnitPriceWithVatCurrency).Multiply(checkoutLine.Quantity)

			newSubtotal, err := subtotal.Add(sub)
			if err != nil {
				errs = errors.Join(errs, err)
			}

			subtotal = newSubtotal
		}

		total, _ := subtotal.Add(deliveryFee)
		total, _ = total.Subtract(totalDiscounts)

		errs = errors.Join(errs, components.CartSummaryRow("Subtotal", subtotal.Display(), "text-gray-500").Render(r.Context(), w))
		errs = errors.Join(errs, components.CartSummaryRow("Total Discount", "- "+totalDiscounts.Display(), "text-red-500").Render(r.Context(), w))
		errs = errors.Join(errs, components.CartSummaryRowWithID("delivery-fee-row", "Delivery Fee", deliveryFee.Display(), "text-gray-500").Render(r.Context(), w))
		errs = errors.Join(errs, components.HR().Render(r.Context(), w))
		errs = errors.Join(errs, components.CartSummaryRow("Total", total.Display()).Render(r.Context(), w))

		if errs != nil {
			logs.Log().Error(logtag, zap.Error(errs))
			http.Error(w, errs.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) getCartLinesCountHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Get Cart Lines Count Handler]"
	set := map[string]bool{}
	if productIDs, ok := s.sessionManager.Get(r.Context(), skCheckoutLineProductIDs).([]string); ok {
		for _, productID := range productIDs {
			if _, exists := set[productID]; !exists {
				set[productID] = true
			}
		}
	}
	if _, err := w.Write(fmt.Appendf(nil, "%d", len(set))); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) updateCartLinesQtyHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Update Cart Lines Qty Handler]"
	checkoutLineID := chi.URLParam(r, "checkoutline_id")
	dbCheckoutLineID := s.encoder.Decode(checkoutLineID)

	qty := 0
	if r.URL.Query().Get("dec") == "1" {
		qty = -1
	} else if r.URL.Query().Get("inc") == "1" {
		qty = 1
	}

	if qty == 0 {
		logs.Log().Error(
			logtag,
			zap.String("checkoutline id", checkoutLineID),
			zap.Error(errs.ErrInvalidParams),
		)
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	newQty, err := s.dbRW.GetQueries().UpdateCheckoutLineQtyByID(
		r.Context(),
		queries.UpdateCheckoutLineQtyByIDParams{
			ID:       dbCheckoutLineID,
			Quantity: int64(qty),
		},
	)
	if err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(fmt.Appendf(nil, "Qty: %d", newQty)); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) cartsFinalizeHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Cart Finalize Handler]"
	var cartCheckout cart.CartCheckout
	if err := utils.FormToStruct(r, &cartCheckout); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := s.sessionManager.Token(r.Context())
	if err := cart.KeepItemsInCheckoutLines(r.Context(), s.dbRW, token, cartCheckout.ToDBCheckoutIDs(s.encoder)); err != nil {
		logs.Log().Error(logtag, zap.Error(err), zap.String("token", token), zap.Any("form", cartCheckout))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	checkoutLines, err := cart.GetCheckoutLines(r.Context(), s.dbRO, token)
	if err != nil {
		logs.Log().Warn(logtag, zap.Error(err), zap.Error(errs.ErrCartMissingCheckoutLines), zap.String("token", token))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch s.paymentGateway.GatewayEnum() {
	case payments.PAYMENT_GATEWAY_PAYMONGO:
		paymentMethods := []payments.PaymentMethod{payments.ParsePaymentMethodToEnum(cartCheckout.PaymentMethod)}
		billing := payments.Billing{
			Address: payments.Address{
				Line1:      cartCheckout.AddressLine1,
				Line2:      cartCheckout.AddressLine2,
				City:       cartCheckout.City,
				State:      cartCheckout.Province,
				PostalCode: cartCheckout.Postal,
				Country:    "PH",
			},
			Name:  cartCheckout.Email,
			Email: cartCheckout.Email,
			Phone: cartCheckout.MobileNo,
		}
		lineItems := make([]payments.LineItem, 0, len(cartCheckout.CheckoutIDs))
		for _, checkoutLine := range checkoutLines {
			lineItems = append(lineItems, payments.LineItem{
				Amount:      int32(checkoutLine.UnitPriceWithVat),
				Currency:    money.PHP,
				Description: checkoutLine.Description.String,
				Images:      []string{s.buildURL(checkoutLine.ThumbnailPath)},
				Name:        checkoutLine.Name,
				Quantity:    int32(checkoutLine.Quantity),
			})
		}

		payload := s.paymentGateway.CreatePayload(billing, lineItems, paymentMethods)
		if err := s.paymentGateway.CheckoutPaymentHandler(w, r, payload); err != nil {
			logs.Log().Error(logtag, zap.Error(err), zap.Any("payload", payload))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		err := fmt.Errorf("%s. %w", logtag, errs.ErrServerUnimplementedGateway)
		logs.Log().Error(err.Error(), zap.String("gateway", s.paymentGateway.GatewayEnum().String()))
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}
}

func (s *Server) cartsPaymentMethodsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Cart Payment Methods Handler]"

	cod, err := s.dbRO.GetQueries().GetSettingsCOD(r.Context())
	paymentMethods := []models.AvailablePaymentMethod{
		{
			Value:     payments.PAYMENT_METHOD_COD,
			Enabled:   (err == nil && cod),
			ImageData: payments.PAYMENT_METHOD_COD.GetImageData(s.cache, s.fs),
		},
	}

	switch s.paymentGateway.GatewayEnum() {
	case payments.PAYMENT_GATEWAY_PAYMONGO:
		availablePaymongoMethods, err := s.paymentGateway.GetAvailablePaymentMethods()
		if err != nil {
			logs.Log().Error(
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
				Value:     pm,
				Enabled:   enabled,
				ImageData: pm.GetImageData(s.cache, s.fs),
			})
		}

	default:
		err := errors.New("checkouts handler. Unimplemented payment gateway")
		logs.Log().Error(err.Error(), zap.String("gateway", s.paymentGateway.GatewayEnum().String()))
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	for _, pm := range paymentMethods {
		if err := components.CartPaymentMethod(pm).Render(r.Context(), w); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
