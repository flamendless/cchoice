package server

import (
	"database/sql"
	"net/http"
	"net/url"
	"strings"

	"cchoice/cmd/web"
	"cchoice/cmd/web/components"
	"cchoice/cmd/web/models"
	"cchoice/internal/database/queries"
	"cchoice/internal/logs"
	"cchoice/internal/requests"
	"cchoice/internal/serialize"
	"cchoice/internal/utils"

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

	r.Route("/cchoice", func(r chi.Router) {
		r.Use(middleware.StripPrefix("/cchoice"))

		// r.Handle("/static/*", http.FileServer(http.FS(web.Files)))
		s.fs = http.FileServer(http.FS(web.Files))
		r.Get("/static/*", s.staticHandler)

		r.Get("/health", s.healthHandler)
		r.Get("/", s.indexHandler)
		r.Get("/settings/header-texts", s.headerTextsHandler)
		r.Get("/settings/footer-texts", s.footerTextsHandler)
		r.Get("/product-categories/side-panel/list", s.categoriesSidePanelHandler)
		r.Get("/product-categories/sections", s.categorySectionHandler)
		r.Get("/product-categories/{category_id}/products", s.categoryProductsHandler)
	})

	return r
}

func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) {
	thumbnail := r.URL.Query().Get("thumbnail")
	if thumbnail == "1" {
		currentURL := r.URL.Path
		if strings.HasPrefix(currentURL, "/static/images/product_images/") {
			currentURL = strings.Replace(currentURL, "/images/", "/thumbnails/", 1)
			newURL, err := url.Parse(currentURL)
			if err != nil {
				panic(err)
			}
			r.URL = newURL
		}
	}
	s.fs.ServeHTTP(w, r)
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
	settings, err := requests.GetSettingsData(
		r.Context(),
		s.Cache,
		&s.SF,
		s.dbRO,
		[]byte("key_header_texts"),
		[]string{"email", "mobile_no"},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("header texts handler", zap.Error(err))
		return
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
	settings, err := requests.GetSettingsData(
		r.Context(),
		s.Cache,
		&s.SF,
		s.dbRO,
		[]byte("key_footer_texts"),
		[]string{"mobile_no", "email", "url_gmap", "url_facebook", "url_tiktok"},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("header texts handler", zap.Error(err))
		return
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

func (s *Server) categoriesSidePanelHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := requests.GetCategoriesSidePanel(
		r.Context(),
		s.Cache,
		&s.SF,
		s.dbRO,
		[]byte("key_categories_side_panel"),
		queries.GetProductCategoriesByPromotedParams{
			PromotedAtHomepage: sql.NullBool{Bool: true, Valid: true},
			Limit:              100,
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("categories side panel list handler", zap.Error(err))
		return
	}

	if err := components.CategoriesSidePanelList(categories).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("categories side panel list handler", zap.Error(err))
	}
}

func (s *Server) categorySectionHandler(w http.ResponseWriter, r *http.Request) {
	res, err := s.dbRO.GetQueries().GetProductCategoriesForSections(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("category section list handler", zap.Error(err))
		return
	}

	categorySections := make([]models.CategorySection, 0, len(res))
	for _, v := range res {
		if v.ProductsCount == 0 {
			logs.Log().Debug(
				"category section has no prododuct. Skipping...",
				zap.String("category name", v.Category.String),
			)
			continue
		}
		categorySections = append(categorySections, models.CategorySection{
			ID:    serialize.EncDBID(v.ID),
			Label: utils.SlugToTile(v.Category.String),
		})
	}

	if err := components.CategorySection(categorySections).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("category section list handler", zap.Error(err))
	}
}

func (s *Server) categoryProductsHandler(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "category_id")
	if categoryID == "" {
		http.Error(w, "Invalid url parameter", http.StatusBadRequest)
		logs.Log().Fatal("category products list handler")
		return
	}

	categoryDBID := serialize.DecDBID(categoryID)

	category, err := s.dbRO.GetQueries().GetProductCategoryByID(r.Context(), categoryDBID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("category section list handler", zap.Error(err))
		return
	}

	if category.Category.String == "" {
		logs.Log().Warn(
			"category has no category value",
			zap.Int64("category id", category.ID),
			zap.String("subcategory", category.Subcategory.String),
		)
		return
	}

	products, err := s.dbRO.GetQueries().GetProductsByCategoryID(r.Context(), queries.GetProductsByCategoryIDParams{
		CategoryID: categoryDBID,
		Limit:      16,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("category section list handler", zap.Error(err))
		return
	}
	if len(products) == 0 {
		logs.Log().Debug(
			"category has no product",
			zap.Int64("category id", category.ID),
			zap.String("category name", category.Category.String),
		)
		return
	}

	categorySectionProducts := models.CategorySectionProducts{
		ID:          categoryID,
		Category:    utils.SlugToTile(category.Category.String),
		Subcategory: utils.SlugToTile(category.Subcategory.String),
		Products:    models.ToCategorySectionProducts(products),
	}

	if err := components.CategorySectionProducts(categorySectionProducts).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logs.Log().Fatal("category section list handler", zap.Error(err))
	}
}
