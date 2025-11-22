package server

import (
	"cchoice/cmd/parse_map/models"
	"cchoice/cmd/web/components"
	"cchoice/internal/cart"
	"cchoice/internal/errs"
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
}

func (s *Server) shippingAddressHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Shipping Address Handler]"

	data := r.URL.Query().Get("data")
	var maps []*models.Map
	switch data {
	case "provinces":
		cachedMaps, err := requests.GetProvinces(s.cache, &s.SF)
		if err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		maps = cachedMaps

	case "cities":
		province := r.URL.Query().Get("province")
		if province == "" {
			logs.Log().Error(logtag, zap.Error(errs.ErrInvalidParams), zap.String("province", province))
			http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
			return
		}

		cachedMaps, err := requests.GetCitiesByProvince(s.cache, &s.SF, province)
		if err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		maps = cachedMaps
		if err := components.MapOption("", "Select City / Municipality").Render(r.Context(), w); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "barangays":
		city := r.URL.Query().Get("city")
		if city == "" {
			logs.Log().Error(logtag, zap.Error(errs.ErrInvalidParams), zap.String("city", city))
			http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
			return
		}

		cachedMaps, err := requests.GetBarangaysByCity(s.cache, &s.SF, city)
		if err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		maps = cachedMaps
		if err := components.MapOption("", "Select Barangay").Render(r.Context(), w); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		logs.Log().Error(logtag, zap.Error(errs.ErrInvalidParams))
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	if len(maps) == 0 {
		err := fmt.Errorf("%w for '%s'", errs.ErrServerNoMapsFound, data)
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	models.SortMap(maps)
	for _, m := range maps {
		if err := components.MapOption(m.Name, m.Name).Render(r.Context(), w); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) shippingQuotationHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Shipping Quotation Handler]"

	token := s.sessionManager.Token(r.Context())
	checkoutLines, err := cart.GetCheckoutLines(r.Context(), s.dbRO, token)
	if err != nil {
		logs.Log().Warn(logtag, zap.Error(err), zap.String("token", token))
		http.Error(w, errs.ErrCartMissingCheckoutLines.Error(), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addressLine1 := strings.TrimSpace(r.Form.Get("address_line1"))
	addressLine2 := strings.TrimSpace(r.Form.Get("address_line2"))
	city := strings.TrimSpace(r.Form.Get("city"))
	province := strings.TrimSpace(r.Form.Get("province"))
	barangay := strings.TrimSpace(r.Form.Get("barangay"))
	postal := strings.TrimSpace(r.Form.Get("postal"))

	if addressLine1 == "" || city == "" || province == "" || barangay == "" {
		logs.Log().Error(logtag, zap.Error(errs.ErrInvalidParams))
		http.Error(w, "Missing required address fields", http.StatusBadRequest)
		return
	}

	address := strings.Join([]string{addressLine1, addressLine2, barangay, city, province, postal, "Philippines"}, ", ")
	address = strings.ReplaceAll(address, ", , ", ", ")

	coordinates, err := requests.GetGeocodingCoordinates(s.cache, &s.SF, s.geocoder, address)
	if err != nil {
		logs.Log().Error(logtag, zap.Error(err), zap.String("address", address))
		http.Error(w, "Failed to geocode address", http.StatusInternalServerError)
		return
	}

	totalWeight, err := utils.CalculateTotalWeightFromCheckoutLines(checkoutLines)
	if err != nil {
		logs.Log().Error(logtag, zap.Error(err))
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
		},
		ServiceType: shipping.SERVICE_TYPE_STANDARD,
	}

	quotation, err := requests.GetShippingQuotation(r.Context(), s.cache, &s.SF, s.shippingService, shippingRequest, s.dbRW)
	if err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logs.Log().Info(logtag, zap.Any("quotation", quotation), zap.String("total_weight", totalWeight))

	s.sessionManager.Put(r.Context(), skShippingQuotation, quotation)
	s.sessionManager.Put(r.Context(), skShippingRequest, shippingRequest)

	if err := components.CartSummaryRowWithID("delivery-fee-row", "Delivery Fee", utils.NewMoney(int64(quotation.Fee*100), quotation.Currency).Display(), "text-gray-500").Render(r.Context(), w); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
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
