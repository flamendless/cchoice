package server

import (
	"cchoice/cmd/parse_map/enums"
	"cchoice/cmd/parse_map/models"
	"cchoice/cmd/web/components"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
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
		maps = models.GetMapsByLevel(models.PhilippinesMap, enums.LEVEL_PROVINCE)

	case "cities":
		province := r.URL.Query().Get("province")
		if province == "" {
			logs.Log().Fatal(logtag, zap.Error(errs.ErrInvalidParams), zap.String("province", province))
			http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
			return
		}
		provinceMap := models.BinarySearchMapByName(models.PhilippinesMap, province, enums.LEVEL_PROVINCE)
		if provinceMap == nil {
			logs.Log().Fatal(logtag, zap.Error(errs.ErrInvalidParams))
			http.Error(w, errs.ErrInvalidParams.Error(), http.StatusInternalServerError)
			return
		}
		maps = provinceMap.Contents
		if err := components.MapOption("", "Select City/Municipality").Render(r.Context(), w); err != nil {
			logs.Log().Fatal(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "barangays":
		city := r.URL.Query().Get("city")
		if city == "" {
			logs.Log().Fatal(logtag, zap.Error(errs.ErrInvalidParams), zap.String("city", city))
			http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
			return
		}
		cityMap := models.BinarySearchMapByName(models.PhilippinesMap, city, enums.LEVEL_CITY)
		if cityMap == nil {
			cityMap = models.BinarySearchMapByName(models.PhilippinesMap, city, enums.LEVEL_MUNICIPALITY)
			if cityMap == nil {
				logs.Log().Fatal(logtag, zap.Error(errs.ErrInvalidParams))
				http.Error(w, errs.ErrInvalidParams.Error(), http.StatusInternalServerError)
				return
			}
		}
		maps = cityMap.Contents
		if err := components.MapOption("", "Select Barangay").Render(r.Context(), w); err != nil {
			logs.Log().Fatal(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		logs.Log().Fatal(logtag, zap.Error(errs.ErrInvalidParams))
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	if len(maps) == 0 {
		err := fmt.Errorf("no maps found for '%s'", data)
		logs.Log().Fatal(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	models.SortMap(maps)
	for _, m := range maps {
		if err := components.MapOption(m.Name, m.Name).Render(r.Context(), w); err != nil {
			logs.Log().Fatal(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
