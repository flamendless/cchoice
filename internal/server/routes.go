package server

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"cchoice/cmd/web/components"
	compfooter "cchoice/cmd/web/components/footer"
	compheader "cchoice/cmd/web/components/header"
	compsearch "cchoice/cmd/web/components/search"
	compshop "cchoice/cmd/web/components/shop"
	"cchoice/cmd/web/models"
	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/enums"
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

func buildImageCacheKey(path, thumbnail, size, quality string, ext enums.ImageFormat) []byte {
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
	imgFormat := enums.ParseImageFormatExtToEnum(ext)
	if imgFormat == enums.IMAGE_FORMAT_UNDEFINED {
		return "", errs.ErrPathInvalidExt
	}

	if len(cleanPath) > constants.MaxImagePathLength {
		return "", errs.ErrPathTooLong
	}

	return cleanPath, nil
}

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestID)
	r.Use(PrometheusMiddleware)
	r.Use(middleware.Logger)
	r.Use(SecurityHeadersMiddleware)
	// r.Use(middleware.NoCache)
	r.Use(middleware.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   conf.Conf().AllowedOrigins(),
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

	r.With(MetricsBasicAuth).Handle("/metrics", promhttp.Handler())
	r.Post("/metrics/event", s.metricsEventHandler)

	r.Get("/", s.indexHandler)
	r.Get("/settings/header-texts", s.headerTextsHandler)
	r.Get("/settings/footer-texts", s.footerTextsHandler)
	r.Get("/settings/store", s.storeHandler)
	r.Get("/products/image", s.productsImageHandler)
	r.Get("/brands/logo", s.brandLogoHandler)
	r.Get("/assets/image", s.assetImageHandler)

	r.Post("/search", s.searchHandler)

	AddProductCategoriesHandlers(s, r)
	AddProductHandlers(s, r)
	AddBrandsHandlers(s, r)
	AddCartsHandlers(s, r)
	AddOrdersHandlers(s, r)
	AddShippingHandlers(s, r)
	AddPaymentHandlers(s, r)
	RegisterPaymentWebhooks(s, r)
	AddAuthHandlers(s, r)
	AddAdminHandlers(s, r)
	AddCustomerHandlers(s, r)
	AddCPointsHandlers(s, r)

	//INFO: (Brandon) - unused routes
	r.Post("/checkouts", s.checkoutsHandler)

	r.Get("/terms", s.termsHandler)
	r.Get("/privacy", s.privacyHandler)
	r.NotFound(s.maintenancePageHandler)
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

	ext := enums.IMAGE_FORMAT_PNG
	thumbnail := r.URL.Query().Get("thumbnail")
	if thumbnail == "1" {
		ext = enums.IMAGE_FORMAT_WEBP
	}

	size := r.URL.Query().Get("size")
	quality := r.URL.Query().Get("quality")
	if quality == "best" {
		size = "640x640"
		ext = enums.IMAGE_FORMAT_WEBP
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

	ext := enums.IMAGE_FORMAT_WEBP
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

	ext := enums.IMAGE_FORMAT_WEBP
	cacheKey := buildImageCacheKey(cleanPath, "", "", "", ext)
	s.serveImage(w, r, cleanPath, ext, cacheKey, logtag)
}

func (s *Server) serveImage(
	w http.ResponseWriter,
	r *http.Request,
	path string,
	ext enums.ImageFormat,
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

	var randomSaleProduct models.RandomSaleProduct
	if conf.Conf().Settings.ShowRandomSaleProduct {
		requestID := middleware.GetReqID(ctx)
		saleProductCacheKey := requests.GenerateRandomSaleProductCacheKey(requestID)
		res, err := requests.GetRandomSaleProduct(
			ctx,
			s.cache,
			&s.SF,
			s.dbRO,
			s.encoder,
			s.GetCDNURL,
			saleProductCacheKey,
		)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("message", "failed to get random sale product"))
		}
		if res != nil {
			randomSaleProduct = *res
			metrics.Promo.ProductImpressionHit(randomSaleProduct.ProductID, randomSaleProduct.Name)
		}
	}

	var promoBanners []models.PromoItem
	if conf.Conf().Settings.ShowPromoBanners {
		activePromos, err := s.services.promo.GetActivePromos(ctx)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err), zap.String("message", "failed to get active promos"))
		}

		promoBanners = make([]models.PromoItem, 0, len(activePromos))
		for _, p := range activePromos {
			promoBanners = append(promoBanners, models.PromoItem{
				ID:          s.encoder.Encode(p.ID),
				Title:       p.Title,
				Description: p.Description,
				MediaURL:    p.MediaURL,
				Type:        p.Type,
			})
		}
	}

	logs.Log().Info(
		logtag,
		zap.String("random sale product", randomSaleProduct.ProductID),
		zap.Int("promo banners", len(promoBanners)),
	)

	homePageData := models.HomePageData{
		Sections:          models.BuildPostHomeContentSections(s.GetBrandLogoCDNURL),
		RandomSaleProduct: randomSaleProduct,
		ActivePromos:      promoBanners,
	}

	if err := compshop.HomePage(homePageData).Render(ctx, w); err != nil {
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
	var parsedAppEnv enums.AppEnv
	if queryAppEnv == "" {
		parsedAppEnv = conf.Conf().AppEnv
	} else {
		parsedAppEnv = enums.ParseAppEnvToEnum(queryAppEnv)
	}

	ctx := r.Context()
	cacheKey := []byte("changelogs:" + parsedAppEnv.String())
	logsData, err := requests.GetChangeLogs(
		ctx,
		s.cache,
		&s.SF,
		cacheKey,
		parsedAppEnv,
		limit,
	)
	if err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Stringer("query", parsedAppEnv),
			zap.Int("limit", limit),
			zap.Error(err),
		)
		http.Error(w, "failed to load changelogs", http.StatusInternalServerError)
		return
	}

	if err := components.ChangeLogs(logsData, limit).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Stringer("query", parsedAppEnv),
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
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResp); err != nil {
		logs.LogCtx(ctx).Error(
			logtag,
			zap.Error(err),
		)
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
	cfg := conf.Conf()
	ctx := r.Context()

	logInLabel := "Log In"

	customerIDStr := s.sessionManager.GetString(ctx, SessionCustomerID)
	if customerIDStr != "" {
		profile, err := s.services.customer.BuildProfile(ctx, customerIDStr)
		if err != nil {
			logs.Log().Warn(logtag, zap.Error(err))
		} else {
			logInLabel = "Hi, " + profile.FirstName
		}
	}

	texts := []models.HeaderRowText{
		{
			Label: "Call Us",
			URL:   constants.ViberURIPrefix + cfg.Settings.MobileNo,
		},
		{
			Label: "E-Mail Us",
			URL:   "mailto:" + cfg.Settings.EMail,
		},
		{
			Label: logInLabel,
			URL:   utils.URL("/customer"),
		},
	}

	if err := compheader.HeaderRow1Texts(texts).Render(ctx, w); err != nil {
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
	cfg := conf.Conf()

	texts := []models.FooterRowText{
		{
			Label: "Home",
			URL:   utils.URL("/"),
		},
		{
			Label: "Admin",
			URL:   utils.URL("/admin"),
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
			URL:   constants.ViberURIPrefix + cfg.Settings.MobileNo,
		},
		{
			Label: "E-Mail Us",
			URL:   "mailto:" + cfg.Settings.EMail,
		},
		{
			Label: "GMaps",
			URL:   cfg.Settings.URLGMap,
		},
		{
			Label: "Waze",
			URL:   cfg.Settings.URLWaze,
		},
		{
			Label:    "Store",
			URL:      utils.URL("#store"),
			Hideable: true,
		},
		{
			Label:    "Facebook",
			URL:      cfg.Settings.URLFacebook,
			Hideable: true,
		},
		{
			Label:    "TikTok",
			URL:      cfg.Settings.URLTikTok,
			Hideable: true,
		},
	}

	ctx := r.Context()
	if err := compfooter.FooterRow1Texts(texts).Render(ctx, w); err != nil {
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
	cfg := conf.Conf()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if _, err := w.Write([]byte(cfg.Settings.Address)); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}

	if err := compshop.GMapsAndWaze(
		cfg.Settings.URLGMap,
		cfg.Settings.URLWaze,
	).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Search Handler]"
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
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
		logs.LogCtx(ctx).Error(
			logtag,
			zap.String("query", searchQuery),
			zap.Error(err),
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
		if err := compsearch.SearchResultProductCard(product).Render(ctx, w); err != nil {
			logs.LogCtx(ctx).Error(
				logtag,
				zap.Error(err),
			)
			return
		}
	}

	if err := compsearch.SearchMore(searchQuery).Render(ctx, w); err != nil {
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

func (s *Server) metricsEventHandler(w http.ResponseWriter, r *http.Request) {
	event := r.URL.Query().Get("event")
	if event == "" {
		http.Error(w, "missing event", http.StatusBadRequest)
		return
	}
	value := r.URL.Query().Get("value")

	logs.Log().Info(
		"Got client event",
		zap.String("event", event),
		zap.String("value", value),
	)

	metrics.ClientEvent.ClientEventHit(event, value)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) maintenancePageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Maintenance Page Handler]"
	ctx := r.Context()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.MaintenancePage().Render(ctx, w); err != nil {
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

func (s *Server) termsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Terms Handler]"
	ctx := r.Context()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.TermsPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) privacyHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Privacy Handler]"
	ctx := r.Context()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.PrivacyPage().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
