package server

import (
	"cchoice/cmd/parse_map/models"
	"cchoice/cmd/web/components"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddShippingHandlers(s *Server, r chi.Router) {
	r.Get("/shipping/address", s.shippingAddressHandler)
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
