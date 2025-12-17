package server

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"cchoice/cmd/web/components"
	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/images"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"
	"cchoice/internal/payments"
	"cchoice/internal/requests"
	"cchoice/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/goccy/go-json"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func buildImageCacheKey(path, thumbnail, size, quality string, ext images.ImageFormat) []byte {
	key := fmt.Sprintf(
		"product_image_%s_t%s_s%s_q%s_%s",
		path,
		thumbnail,
		size,
		quality,
		ext.String(),
	)
	return []byte(key)
}

func validateImagePath(path string, allowedPrefixes []string) (string, error) {
	if path == "" {
		return "", errs.ErrPathEmpty
	}

	if strings.Contains(path, "\x00") {
		return "", errs.ErrPathEmpty
	}

	cleanPath := filepath.Clean(path)

	if strings.Contains(cleanPath, "..") {
		return "", errs.ErrPathTraversalAttempt
	}

	hasValidPrefix := false
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(cleanPath, prefix) {
			hasValidPrefix = true
			break
		}
	}
	if !hasValidPrefix {
		return "", errs.ErrPathPrefix
	}

	ext := filepath.Ext(cleanPath)
	imgFormat := images.ParseImageFormatExtToEnum(ext)
	if imgFormat == images.IMAGE_FORMAT_UNDEFINED {
		return "", errs.ErrPathInvalidExt
	}

	if len(cleanPath) > 512 {
		return "", errs.ErrPathTooLong
	}

	return cleanPath, nil
}

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(SecurityHeadersMiddleware)
	// r.Use(middleware.NoCache)
	r.Use(middleware.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	urlPrefix := utils.URL("")
	if urlPrefix != "" {
		r.Route(urlPrefix, func(r chi.Router) {
			r.Use(middleware.StripPrefix(urlPrefix))
			s.registerAllRoutes(r)
		})
	} else {
		s.registerAllRoutes(r)
	}

	return r
}

func (s *Server) registerAllRoutes(r chi.Router) {
	// Use staticFS for static assets (JS, CSS, icons)
	if s.staticFS == nil {
		panic(errors.Join(errs.ErrServerInit, errs.ErrServerFSNotSetup))
	}

	// Custom static file handler with caching
	r.Handle("/static/*", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		const logtag = "[Static Handler]"
		ctx := req.Context()

		path := req.URL.Path
		notModified, file, err := httputil.CacheHeaders(w, req, s.staticFS, path)
		if err != nil {
			logs.LogCtx(ctx).Debug(
				logtag,
				zap.String("path", path),
				zap.Error(err),
			)
			httputil.SetNoCacheHeaders(w)
			http.NotFound(w, req)
			return
		}
		defer file.Close()

		if notModified {
			return
		}

		info, err := file.Stat()
		if err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			httputil.SetNoCacheHeaders(w)
			http.NotFound(w, req)
			return
		}

		http.ServeContent(w, req, info.Name(), info.ModTime(), file)
	})))

	r.Get("/changelogs", s.changelogsHandler)
	r.Get("/health", s.healthHandler)
	r.Get("/version", s.versionHandler)
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/", s.indexHandler)
	r.Get("/settings/header-texts", s.headerTextsHandler)
	r.Get("/settings/footer-texts", s.footerTextsHandler)
	r.Get("/settings/store", s.storeHandler)
	r.Get("/products/image", s.productsImageHandler)
	r.Get("/brands/logo", s.brandLogoHandler)
	r.Get("/assets/image", s.assetImageHandler)

	r.Post("/search", s.searchHandler)

	AddProductCategoriesHandlers(s, r)
	AddBrandsHandlers(s, r)
	AddCartsHandlers(s, r)
	AddShippingHandlers(s, r)
	AddPaymentHandlers(s, r)
	RegisterPaymentWebhooks(s, r)

	//INFO: (Brandon) - unused routes
	r.Post("/checkouts", s.checkoutsHandler)
}

func (s *Server) productsImageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Products Image Handler]"
	ctx := r.Context()

	path := r.URL.Query().Get("path")

	cleanPath, err := validateImagePath(path, []string{constants.PathProductImages})
	if err != nil {
		logs.LogCtx(ctx).Warn(
			logtag,
			zap.String("path", path),
			zap.Error(err),
		)
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	ext := images.IMAGE_FORMAT_PNG
	thumbnail := r.URL.Query().Get("thumbnail")
	if thumbnail == "1" {
		ext = images.IMAGE_FORMAT_WEBP
	}

	size := r.URL.Query().Get("size")
	quality := r.URL.Query().Get("quality")
	if quality == "best" {
		size = "640x640"
		ext = images.IMAGE_FORMAT_WEBP
	}

	cacheKey := buildImageCacheKey(cleanPath, thumbnail, size, quality, ext)
	s.serveImage(w, r, cleanPath, ext, cacheKey, logtag)
}

func (s *Server) brandLogoHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Brand Logo Handler]"
	ctx := r.Context()

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		logs.LogCtx(ctx).Debug(
			logtag,
			zap.String("error", "missing filename parameter"),
		)
		http.Error(w, "missing filename parameter", http.StatusBadRequest)
		return
	}

	path := "static/images/brand_logos/" + filename

	cleanPath, err := validateImagePath(path, []string{"static/images/brand_logos/"})
	if err != nil {
		logs.LogCtx(ctx).Warn(
			logtag,
			zap.String("path", path),
			zap.Error(err),
		)
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	ext := images.IMAGE_FORMAT_WEBP
	cacheKey := buildImageCacheKey(cleanPath, "", "", "", ext)
	s.serveImage(w, r, cleanPath, ext, cacheKey, logtag)
}

func (s *Server) assetImageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Asset Image Handler]"
	ctx := r.Context()

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		logs.LogCtx(ctx).Debug(
			logtag,
			zap.String("error", "missing filename parameter"),
		)
		http.Error(w, "missing filename parameter", http.StatusBadRequest)
		return
	}

	path := "static/images/" + filename

	cleanPath, err := validateImagePath(path, []string{"static/images/"})
	if err != nil {
		logs.LogCtx(ctx).Warn(
			logtag,
			zap.String("path", path),
			zap.Error(err),
		)
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	ext := images.IMAGE_FORMAT_WEBP
	cacheKey := buildImageCacheKey(cleanPath, "", "", "", ext)
	s.serveImage(w, r, cleanPath, ext, cacheKey, logtag)
}

func (s *Server) serveImage(
	w http.ResponseWriter,
	r *http.Request,
	path string,
	ext images.ImageFormat,
	cacheKey []byte,
	logtag string,
) {
	ctx := r.Context()
	if data, ok := s.cache.HasGet(nil, cacheKey); ok {
		w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := w.Write(data); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		metrics.Cache.MemHit()
		return
	} else {
		metrics.Cache.MemMiss()
	}

	if notModified, file, err := httputil.CacheHeaders(w, r, s.productImageFS, path); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		httputil.SetNoCacheHeaders(w)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else {
		defer file.Close()
		if notModified {
			return
		}
	}

	imgData, err := images.GetImageDataB64(s.cache, s.productImageFS, path, ext)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", path),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte(imgData)); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.cache.Set(cacheKey, []byte(imgData))
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Index handler]"
	ctx := r.Context()

	homePageData := models.HomePageData{
		Sections:      models.BuildPostHomeContentSections(s.GetBrandLogoCDNURL),
		StoreImageURL: s.GetAssetCDNURL("store.webp"),
	}

	if err := components.HomePage(homePageData).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) changelogsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Changelogs Handler]"
	const limit = 8

	queryAppEnv := r.URL.Query().Get("appenv")
	if queryAppEnv == "" {
		queryAppEnv = conf.Conf().AppEnv
	}

	ctx := r.Context()
	cacheKey := []byte("changelogs:" + queryAppEnv)
	logsData, err := requests.GetChangeLogs(
		ctx,
		s.cache,
		&s.SF,
		cacheKey,
		queryAppEnv,
		limit,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("query", queryAppEnv),
			zap.Int("limit", limit),
			zap.Error(err),
		)
		http.Error(w, "failed to load changelogs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.ChangeLogs(logsData, limit).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("query", queryAppEnv),
			zap.Int("limit", limit),
			zap.Error(err),
		)
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Health Handler]"
	ctx := r.Context()

	jsonResp, err := json.Marshal(s.dbRO.Health())
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(jsonResp); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) versionHandler(w http.ResponseWriter, r *http.Request) {
	ver := conf.GitTagProd
	if conf.Conf().IsLocal() {
		ver = conf.GitTagDev
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(ver))
}

func (s *Server) headerTextsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Header Texts Handler]"
	ctx := r.Context()

	settings, err := requests.GetSettingsData(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		[]byte("key_header_texts"),
		[]string{"email", "mobile_no"},
	)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
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

	if err := components.HeaderRow1Texts(texts).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) footerTextsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Footer Texts Handler]"
	ctx := r.Context()

	settings, err := requests.GetSettingsData(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		[]byte("key_footer_texts"),
		[]string{
			"mobile_no",
			"email",
			"address",
			"url_gmap",
			"url_facebook",
			"url_tiktok",
		},
	)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	texts := []models.FooterRowText{
		{
			Label: "Home",
			URL:   utils.URL("/"),
		},
		{
			Label:    "About Us",
			URL:      utils.URL("#about-us"),
			Hideable: true,
		},
		{
			Label:    "Services",
			URL:      utils.URL("#services"),
			Hideable: true,
		},
		{
			Label: "Partners",
			URL:   utils.URL("#partners"),
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
			Label:    "Store",
			URL:      utils.URL("#store"),
			Hideable: true,
		},
		{
			Label:    "Facebook",
			URL:      settings["url_facebook"],
			Hideable: true,
		},
		{
			Label:    "TikTok",
			URL:      settings["url_tiktok"],
			Hideable: true,
		},
	}

	if err := components.FooterRow1Texts(texts).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) storeHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Store Handler]"
	ctx := r.Context()
	settings, err := requests.GetSettingsData(
		ctx,
		s.cache,
		&s.SF,
		s.dbRO,
		[]byte("key_store"),
		[]string{"address"},
	)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if addr, ok := settings["address"]; ok {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(addr))
	}
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Search Handler]"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	search := r.PostFormValue("search")
	searchMobile := r.PostFormValue("search-mobile")
	searchQuery := cmp.Or(search, searchMobile)

	products, err := s.dbRO.GetQueries().GetProductsBySearchQuery(
		ctx,
		queries.GetProductsBySearchQueryParams{
			Name:  searchQuery,
			Limit: constants.MaxSearchShowResults,
		},
	)
	if err != nil || len(products) == 0 {
		logs.LogCtx(ctx).Info(
			logtag,
			zap.String("query", searchQuery),
		)
		return
	}

	logs.LogCtx(ctx).Info(
		logtag,
		zap.Int("count", len(products)),
		zap.Int("limit", constants.MaxSearchShowResults),
		zap.String("query", searchQuery),
	)

	productResults := make([]models.SearchResultProduct, 0, len(products))
	for i := range products {
		if strings.HasSuffix(products[i].ThumbnailPath, constants.EmptyImageFilename) {
			continue
		}
		productResults = append(productResults, models.ToSearchResultProduct(s.encoder, s.GetCDNURL, products[i]))
	}

	for _, product := range productResults {
		if err := components.SearchResultProductCard(product).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			return
		}
	}

	if err := components.SearchMore(searchQuery).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		return
	}
}

func (s *Server) checkoutsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Checkouts Handler]"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch s.paymentGateway.GatewayEnum() {
	case payments.PAYMENT_GATEWAY_PAYMONGO:
		// if err := s.paymentGateway.CheckoutPaymentHandler(w, r); err != nil {
		// 	logs.LogCtx(ctx).Error("[PayMongo] Checkouts handler", zap.Error(err))
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
	default:
		err := errs.ErrServerUnimplementedGateway
		logs.LogCtx(ctx).Error(
			err.Error(),
			zap.String("gateway", s.paymentGateway.GatewayEnum().String()),
		)
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}
}
