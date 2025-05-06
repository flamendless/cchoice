package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"cchoice/cmd/web"
	"cchoice/cmd/web/components"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/images"
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
	r.Use(middleware.CleanPath)
	r.Use(middleware.Logger)
	r.Use(middleware.NoCache)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.Compress(5))

	r.Route("/cchoice", func(r chi.Router) {
		r.Use(middleware.StripPrefix("/cchoice"))

		// r.Handle("/static/*", http.FileServer(http.FS(web.Files)))
		s.fs = http.FS(web.Files)
		s.fsHandler = http.FileServer(s.fs)
		r.Get("/static/*", s.staticHandler)

		r.Get("/health", s.healthHandler)
		r.Get("/", s.indexHandler)
		r.Get("/settings/header-texts", s.headerTextsHandler)
		r.Get("/settings/footer-texts", s.footerTextsHandler)
		r.Get("/product-categories/side-panel/list", s.categoriesSidePanelHandler)
		r.Get("/product-categories/sections", s.categorySectionHandler)
		r.Get("/product-categories/{category_id}/products", s.categoryProductsHandler)
		r.Get("/products/image", s.productsImageHandler)
	})

	return r
}

func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) {
	s.fsHandler.ServeHTTP(w, r)
}

func (s *Server) productsImageHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" || !strings.HasPrefix(path, constants.PathProductImages) {
		logs.Log().Debug("invalid image prefix", zap.String("path", path))
		return
	}

	key := "product_image_" + path
	isThumbnail := false
	size := r.URL.Query().Get("size")
	if size == "" {
		size = "160x160"
		isThumbnail = true
		key += size
	}

	cacheKey := []byte(key)
	if data, ok := s.Cache.HasGet(nil, cacheKey); ok {
		if err := components.Image(string(data)).Render(r.Context(), w); err != nil {
			logs.Log().Fatal("Product Image handler", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		logs.Log().Debug("cache hit", zap.ByteString("key", cacheKey))
		return
	} else {
		logs.Log().Debug("cache miss", zap.ByteString("key", cacheKey))
	}

	finalPath, ext, err := images.GetImagePathWithSize(path, size, isThumbnail)
	if err != nil {
		logs.Log().Fatal("Product Image handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	imgData, err := images.GetImageDataB64(s.Cache, s.fs, finalPath, ext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := components.Image(imgData).Render(r.Context(), w); err != nil {
		logs.Log().Fatal("Product Image handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.Cache.Set(cacheKey, []byte(imgData))
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := components.Base("C-CHOICE").Render(r.Context(), w); err != nil {
		logs.Log().Fatal("Index handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.dbRO.Health())
	if err != nil {
		logs.Log().Fatal("Health handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		logs.Log().Fatal("header texts handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		logs.Log().Fatal("header texts handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		logs.Log().Fatal("header texts handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	texts := []models.FooterRowText{
		{
			Label: "Home",
			URL:   "/cchoice/",
		},
		{
			Label: "Services",
			URL:   "/cchoice#services",
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
		logs.Log().Fatal("footer texts handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		logs.Log().Fatal("categories side panel list handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := components.CategoriesSidePanelList(categories).Render(r.Context(), w); err != nil {
		logs.Log().Fatal("categories side panel list handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) categorySectionHandler(w http.ResponseWriter, r *http.Request) {
	page := 0
	if paramPage := r.URL.Query().Get("page"); paramPage != "" {
		if parsed, err := strconv.Atoi(paramPage); err == nil {
			page = parsed
		}
	}

	limit := constants.DefaultLimitCategories
	if paramLimit := r.URL.Query().Get("limit"); paramLimit != "" {
		if parsed, err := strconv.Atoi(paramLimit); err == nil {
			limit = max(parsed, constants.DefaultLimitCategories)
		}
	}

	res, err := requests.GetCategorySectionHandler(
		r.Context(),
		s.Cache,
		&s.SF,
		s.dbRO,
		fmt.Appendf([]byte{}, "categorySectionHandler_p%d_l%d", page, limit),
		page,
		limit,
	)
	if err != nil {
		logs.Log().Fatal("category section handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := components.CategorySection(page, res).Render(r.Context(), w); err != nil {
		logs.Log().Fatal("category section list handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) categoryProductsHandler(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "category_id")
	if categoryID == "" {
		logs.Log().Fatal("category products list handler")
		http.Error(w, "Invalid url parameter", http.StatusBadRequest)
		return
	}

	categoryDBID := serialize.DecDBID(categoryID)

	category, err := s.dbRO.GetQueries().GetProductCategoryByID(r.Context(), categoryDBID)
	if err != nil {
		logs.Log().Fatal("category section list handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		Limit:      constants.DefaultLimitProducts,
	})
	if err != nil {
		logs.Log().Fatal("category section list handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	for i, product := range products {
		if product.ThumbnailPath == constants.PathEmptyImage {
			continue
		}

		finalPath, ext, err := images.GetImagePathWithSize(product.ThumbnailPath, "96x96", true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			continue
		}

		imgData, err := images.GetImageDataB64(s.Cache, s.fs, finalPath, ext)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			continue
		}

		products[i].ThumbnailData = imgData
	}

	categorySectionProducts := models.CategorySectionProducts{
		ID:          categoryID,
		Category:    utils.SlugToTile(category.Category.String),
		Subcategory: utils.SlugToTile(category.Subcategory.String),
		Products:    models.ToCategorySectionProducts(products),
	}

	if err := components.CategorySectionProducts(categorySectionProducts).Render(r.Context(), w); err != nil {
		logs.Log().Fatal("category section list handler", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
