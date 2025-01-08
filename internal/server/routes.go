package server

import (
	"encoding/json"
	"log"
	"net/http"

	"cchoice/cmd/web"
	"cchoice/cmd/web/components"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	r.Get("/", s.indexHandler)

	r.Get("/health", s.healthHandler)

	fileServer := http.FileServer(http.FS(web.Files))
	r.Handle("/static/*", fileServer)
	// r.Get("/web", templ.Handler(web.HelloForm()).ServeHTTP)
	// r.Post("/hello", web.HelloWebHandler)

	return r
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := components.Base("C-CHOICE").Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in index handler: %e", err)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.dbRO.Health())
	_, _ = w.Write(jsonResp)
}
