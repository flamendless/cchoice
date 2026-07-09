package server

import (
	"cchoice/cmd/parse_map/models"
	compcart "cchoice/cmd/web/components/cart"
	"cchoice/internal/cart"
	"cchoice/internal/errs"
	"cchoice/internal/geocoding"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/server/forms"
	"cchoice/internal/shipping"
	"cchoice/internal/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddShippingHandlers(s *Server, r chi.Router) {
	r.Get("/shipping/address", s.shippingAddressHandler)
	r.Get("/shipping/quotation/status", s.shippingQuotationStatusHandler)
	r.Post("/shipping/quotation", s.shippingQuotationHandler)
	r.Delete("/shipping/quotation", s.clearShippingQuotationHandler)
}

func (s *Server) shippingAddressHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Shipping Address Handler]"
	ctx := r.Context()

	var req forms.ShippingAddressQuery
	if err := httputil.BindQuery(r, &req); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, httputil.ErrorMessage(err), http.StatusBadRequest)
		return
	}

	var maps []*models.Map
	switch req.Data {
	case "provinces":
		cachedMaps, err := requests.GetProvinces(s.cache, &s.SF)
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		maps = cachedMaps

	case "cities":
		cachedMaps, err := requests.GetCitiesByProvince(s.cache, &s.SF, req.Province)
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		maps = cachedMaps

	case "barangays":
		cachedMaps, err := requests.GetBarangaysByCity(s.cache, &s.SF, req.City)
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		maps = cachedMaps

	default:
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(errs.ErrInvalidParams),
		)
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	if len(maps) == 0 {
		err := fmt.Errorf("%w for '%s'", errs.ErrServerNoMapsFound, req.Data)
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	models.SortMap(maps)
	for _, m := range maps {
		if err := compcart.MapOption(m.Name, m.Name).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) shippingQuotationHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Shipping Quotation Handler]"
	ctx := r.Context()

	token := s.sessionManager.Token(ctx)
	checkoutLines, err := cart.GetCheckoutLines(ctx, s.dbRO, token)
	if err != nil {
		logs.LogCtx(ctx).Warn(
			logtag,
			zap.String("token", token),
			zap.Error(err),
		)
		http.Error(w, errs.ErrCartMissingCheckoutLines.Error(), http.StatusBadRequest)
		return
	}

	var formReq forms.ShippingQuotationForm
	if err := httputil.BindForm(r, &formReq); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, "Missing required address fields (city, province, barangay)", http.StatusBadRequest)
		return
	}

	addressLine1 := formReq.AddressLine1
	addressLine2 := formReq.AddressLine2
	city := formReq.City
	province := formReq.Province
	barangay := formReq.Barangay
	postal := formReq.Postal

	allParts := []string{addressLine1, addressLine2, barangay, city, province, postal, "Philippines"}
	addressParts := make([]string, 0, len(allParts))
	for _, part := range allParts {
		if part != "" {
			addressParts = append(addressParts, part)
		}
	}
	address := strings.Join(addressParts, ", ")

	coordinates, err := requests.GetGeocodingCoordinates(s.cache, &s.SF, s.geocoder, address)
	fallbackSF := false
	if err != nil {
		logs.LogCtx(ctx).Warn(
			logtag,
			zap.String("address", address),
			zap.Error(err),
		)
		coordinates = &geocoding.Coordinates{
			Lat: "0.0",
			Lng: "0.0",
		}
		fallbackSF = true
	}

	totalWeight, err := utils.CalculateTotalWeightFromCheckoutLines(checkoutLines)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, "Failed to calculate package weight", http.StatusInternalServerError)
		return
	}

	businessLocation := s.shippingService.GetBusinessLocation()

	shippingRequest := shipping.ShippingRequest{
		Package: shipping.Package{
			Weight:      totalWeight,
			Description: "Order package",
		},
		PickupLocation: *businessLocation,
		DeliveryLocation: shipping.Location{
			Coordinates: shipping.Coordinates{
				Lat: coordinates.Lat,
				Lng: coordinates.Lng,
			},
			Address: address,
			OriginalAddress: shipping.Address{
				Line1:      addressLine1,
				Line2:      addressLine2,
				City:       city,
				State:      province,
				PostalCode: postal,
				Country:    "PH",
			},
		},
		ServiceType: shipping.SERVICE_TYPE_STANDARD,
	}

	quotation, err := requests.GetShippingQuotation(ctx, s.cache, &s.SF, s.shippingService, shippingRequest, s.dbRW)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	isFreeDelivery, ok := quotation.Metadata["free_delivery"].(bool)
	if fallbackSF && (!ok || !isFreeDelivery) {
		quotation.Fee = 100
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Any("quotation", quotation),
		zap.String("total_weight", totalWeight),
	)

	s.sessionManager.Put(ctx, skShippingQuotation, quotation)
	s.sessionManager.Put(ctx, skShippingRequest, shippingRequest)

	if err := compcart.CartSummaryRowWithID("delivery-fee-row", "Delivery Fee", utils.NewMoney(int64(quotation.Fee*100), quotation.Currency).Display(), "text-gray-500").Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) shippingQuotationStatusHandler(w http.ResponseWriter, r *http.Request) {
	exists := s.sessionManager.Exists(r.Context(), skShippingQuotation)
	if exists {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Server) clearShippingQuotationHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Clear Shipping Quotation Handler]"
	ctx := r.Context()

	s.sessionManager.Remove(ctx, skShippingQuotation)
	s.sessionManager.Remove(ctx, skShippingRequest)

	logs.LogCtx(ctx).Info(
		logtag,
		zap.String("action", "cleared shipping quotation and request from session"),
	)
	w.WriteHeader(http.StatusOK)
}
