package server

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"cchoice/cmd/web"
	"cchoice/cmd/web/components"
	"cchoice/cmd/web/models"
	"cchoice/internal/database/queries"
	"cchoice/internal/logs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	r.Get("/settings/header-texts", s.headerTextsHandler)
	r.Get("/settings/footer-texts", s.footerTextsHandler)
	r.Get("/product-categories/list", s.categoriesListHandler)
	r.Get("/product-categories/subcategories/list", s.categoriesSubcategoriesListHandler)
	r.Get("/health", s.healthHandler)

	return r
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := components.Base("C-CHOICE").Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func (s *Server) headerTextsHandler(w http.ResponseWriter, r *http.Request) {
	res, err := s.dbRO.GetQueries().GetSettingsByNames(
		context.TODO(),
		[]string{"email", "mobile_no"},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("header texts handler", zap.Error(err))
		return
	}

	settings := make(map[string]string, len(res))
	for _, setting := range res {
		settings[setting.Name] = setting.Value
	}

	texts := []models.HeaderRowText{
		{
			Label: "Call Us",
			URL:   "viber://chat?number=" + settings["mobile_no"],
		},
		{
			Label: "E-Mail Us",
			URL:   "mailto:" + settings["email"],
		},
		{
			Label: "Log In",
			URL:   "/log-in",
		},
	}

	if err := components.HeaderRow1Texts(texts).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("header texts handler", zap.Error(err))
	}
}

func (s *Server) footerTextsHandler(w http.ResponseWriter, r *http.Request) {
	res, err := s.dbRO.GetQueries().GetSettingsByNames(
		context.TODO(),
		[]string{
			"mobile_no",
			"email",
			"url_gmap",
			"url_facebook",
			"url_tiktok",
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("footer texts handler", zap.Error(err))
		return
	}

	settings := make(map[string]string, len(res))
	for _, setting := range res {
		settings[setting.Name] = setting.Value
	}

	texts := []models.FooterRowText{
		{
			Label: "Home",
			URL:   "/",
		},
		{
			Label: "Call Us",
			URL:   "viber://chat?number=" + settings["mobile_no"],
		},
		{
			Label: "E-Mail Us",
			URL:   "mailto:" + settings["email"],
		},
		{
			Label: "Location",
			URL:   settings["url_gmap"],
		},
		{
			Label: "Facebook",
			URL:   settings["url_facebook"],
		},
		{
			Label: "TikTok",
			URL:   settings["url_tiktok"],
		},
	}

	if err := components.FooterRow1Texts(texts).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("footer texts handler", zap.Error(err))
	}
}

func (s *Server) categoriesListHandler(w http.ResponseWriter, r *http.Request) {
	res, err := s.dbRO.GetQueries().GetProductCategoriesByPromoted(
		context.TODO(),
		queries.GetProductCategoriesByPromotedParams{
			PromotedAtHomepage: sql.NullBool{Bool: true, Valid: true},
			Limit:              100,
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("categories list handler", zap.Error(err))
		return
	}

	categories := make([]models.CategorySidePanelText, 0, len(res))
	caser := cases.Title(language.English)
	for _, v := range res {
		name := v.Category.String
		keywords := strings.Split(name, "-")
		name = strings.Join(keywords, " ")
		name = caser.String(name)

		categories = append(categories, models.CategorySidePanelText{
			Label: name,
			URL:   "/product-category/" + v.Category.String,
		})
	}

	if err := components.CategoriesList(categories).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("categories list handler", zap.Error(err))
	}
}

func (s *Server) categoriesSubcategoriesListHandler(w http.ResponseWriter, r *http.Request) {
	res, err := s.dbRO.GetQueries().GetProductCategoriesByPromoted(
		context.TODO(),
		queries.GetProductCategoriesByPromotedParams{
			PromotedAtHomepage: sql.NullBool{Bool: true, Valid: true},
			Limit:              100,
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("subcategories list handler", zap.Error(err))
		return
	}

	subcategories := make([]models.Subcategory, 0, len(res))
	caser := cases.Title(language.English)
	for _, v := range res {
		name := v.Category.String
		keywords := strings.Split(name, "-")
		name = strings.Join(keywords, " ")
		name = caser.String(name)

		subcategories = append(subcategories, models.Subcategory{
			Label: name,
		})
	}

	if err := components.SubcategoriesList(subcategories).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("subcategories list handler", zap.Error(err))
	}
}
