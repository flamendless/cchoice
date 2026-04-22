package server

import (
	"net/http"
	"strings"

	compshop "cchoice/cmd/web/components/shop"
	"cchoice/internal/logs"
	"cchoice/internal/requests"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func AddBrandsHandlers(s *Server, r chi.Router) {
	r.Get("/brands/side-panel/list", s.brandsSidePanelHandler)
}

func (s *Server) brandsSidePanelHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Brands Side Panel Handler]"
	ctx := r.Context()
	brands, err := requests.GetBrandsSidePanel(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		s.encoder,
		[]byte("key_brands_side_panel"),
	)
	if err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brandLabel := compshop.PrimaryBrand

	filters := GetHomePageFilters(ctx, s.sessionManager)
	if filters.BrandID != "" {
		brandName, err := s.services.brand.GetNameByID(ctx, filters.BrandID)
		if err != nil {
			logs.Log().Warn(logtag, zap.Error(err), zap.String("brand id", filters.BrandID))
		} else {
			brandLabel = brandName
		}
	}

	if err := compshop.BrandsSidePanelList(strings.ToUpper(brandLabel), brands).Render(ctx, w); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
