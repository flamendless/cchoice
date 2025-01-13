package server

import (
	"net/http"

	"cchoice/cmd/web"
	"cchoice/cmd/web/components"
	"cchoice/internal/logs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Handle("/static/*", http.FileServer(http.FS(web.Files)))

	r.Get("/", s.indexHandler)
	r.Get("/health", s.healthHandler)
	// r.Get("/web", templ.Handler(web.HelloForm()).ServeHTTP)
	// r.Post("/hello", web.HelloWebHandler)

	return r
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := components.Base("C-CHOICE").Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logs.Log().Fatal("Index handler", zap.Error(err))
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.dbRO.Health())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("Health handler", zap.Error(err))
		return
	}
	_, _ = w.Write(jsonResp)
}
