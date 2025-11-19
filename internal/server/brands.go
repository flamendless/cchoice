package server

import (
	"net/http"

	"cchoice/cmd/web/components"
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
	brands, err := requests.GetBrandsSidePanel(
		r.Context(),
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

	if err := components.BrandsSidePanelList(brands).Render(r.Context(), w); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
