package server

import (
	"cchoice/cmd/parse_map/models"
	"cchoice/cmd/web/components"
	"cchoice/internal/cart"
	"cchoice/internal/errs"
	"cchoice/internal/geocoding"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
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

	data := r.URL.Query().Get("data")
	var maps []*models.Map
	switch data {
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
		province := r.URL.Query().Get("province")
		if province == "" {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.String("province", province),
				zap.Error(errs.ErrInvalidParams),
			)
			http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
			return
		}

		cachedMaps, err := requests.GetCitiesByProvince(s.cache, &s.SF, province)
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		maps = cachedMaps
		// Don't render default option here - it's already set by client-side reset handler

	case "barangays":
		city := r.URL.Query().Get("city")
		if city == "" {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.String("city", city),
				zap.Error(errs.ErrInvalidParams),
			)
			http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
			return
		}

		cachedMaps, err := requests.GetBarangaysByCity(s.cache, &s.SF, city)
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		maps = cachedMaps
		// Don't render default option here - it's already set by client-side reset handler

	default:
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(errs.ErrInvalidParams),
		)
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	if len(maps) == 0 {
		err := fmt.Errorf("%w for '%s'", errs.ErrServerNoMapsFound, data)
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	models.SortMap(maps)
	for _, m := range maps {
		if err := components.MapOption(m.Name, m.Name).Render(ctx, w); err != nil {
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

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addressLine1 := strings.TrimSpace(r.Form.Get("address_line1"))
	addressLine2 := strings.TrimSpace(r.Form.Get("address_line2"))
	city := strings.TrimSpace(r.Form.Get("city"))
	province := strings.TrimSpace(r.Form.Get("province"))
	barangay := strings.TrimSpace(r.Form.Get("barangay"))
	postal := strings.TrimSpace(r.Form.Get("postal"))

	if province == "National Capital Region (NCR)" {
		fmt.Println("NCR province selected")
	} else if province == "" || city == "" || barangay == "" {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(errs.ErrInvalidParams),
			zap.String("province", province),
			zap.String("city", city),
			zap.String("barangay", barangay),
		)
		http.Error(w, "Missing required address fields (city, province, barangay)", http.StatusBadRequest)
		return
	}

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

	if err := components.CartSummaryRowWithID("delivery-fee-row", "Delivery Fee", utils.NewMoney(int64(quotation.Fee*100), quotation.Currency).Display(), "text-gray-500").Render(ctx, w); err != nil {
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
