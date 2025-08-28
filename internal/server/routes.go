package server

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"cchoice/cmd/web/components"
	"cchoice/cmd/web/models"
	"cchoice/cmd/web/static"
	"cchoice/internal/constants"
	"cchoice/internal/database/queries"
	"cchoice/internal/errs"
	"cchoice/internal/images"
	"cchoice/internal/logs"
	"cchoice/internal/metrics"
	"cchoice/internal/payments"
	"cchoice/internal/requests"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/goccy/go-json"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.CleanPath)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	// r.Use(middleware.NoCache)
	r.Use(middleware.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/cchoice", func(r chi.Router) {
		r.Use(middleware.StripPrefix("/cchoice"))

		fs := static.GetFS()
		if fs == nil {
			panic(errors.Join(errs.ErrServerInit, errors.New("server.fs not setup")))
		}

		s.fs = http.FS(fs)
		r.Handle(
			"/static/*",
			// http.StripPrefix("/static/", static.Handler()),
			http.StripPrefix("/static/", static.CacheHandler(CacheHeaders)),
		)

		r.Get("/changelogs", s.changelogsHandler)
		r.Get("/health", s.healthHandler)
		r.Handle("/metrics", promhttp.Handler())
		r.Get("/", s.indexHandler)
		r.Get("/settings/header-texts", s.headerTextsHandler)
		r.Get("/settings/footer-texts", s.footerTextsHandler)
		r.Get("/products/image", s.productsImageHandler)

		r.Post("/search", s.searchHandler)

		AddProductCategoriesHandlers(s, r)
		AddCartsHandlers(s, r)
		AddShippingHandlers(s, r)

		//INFO: (Brandon) - unused routes
		r.Post("/checkouts", s.checkoutsHandler)
	})

	return r
}

func (s *Server) productsImageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Products Image Handler]"
	path := r.URL.Query().Get("path")
	if path == "" || !strings.HasPrefix(path, constants.PathProductImages) {
		logs.Log().Debug(
			logtag,
			zap.Error(errs.ErrImagePrefix),
			zap.String("path", path),
		)
		return
	}

	key := "product_image_" + path
	isThumbnail := false
	thumbnail := r.URL.Query().Get("thumbnail")
	if thumbnail == "1" {
		isThumbnail = true
	}

	size := r.URL.Query().Get("size")
	if size == "" {
		size = "160x160"
		isThumbnail = true
		key += size
	}

	cacheKey := []byte(key)
	if data, ok := s.cache.HasGet(nil, cacheKey); ok {
		if err := components.Image(string(data)).Render(r.Context(), w); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		metrics.Cache.MemHit()
		return
	} else {
		metrics.Cache.MemMiss()
	}

	finalPath, ext, err := images.GetImagePathWithSize(path, size, isThumbnail)
	if err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if notModified, file, err := CacheHeaders(w, r, s.fs, finalPath); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else {
		defer file.Close()
		if notModified {
			return
		}
	}

	imgData, err := images.GetImageDataB64(s.cache, s.fs, finalPath, ext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := components.Image(imgData).Render(r.Context(), w); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.cache.Set(cacheKey, []byte(imgData))
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Index handler]"
	if err := components.HomePage().Render(r.Context(), w); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) changelogsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Changelogs Handler]"
	f, err := os.Open("./CHANGELOGS.md")
	if err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}()

	w.Header().Set("Content-Type", "text/plain")
	if _, err := io.Copy(w, f); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Health Handler]"
	jsonResp, err := json.Marshal(s.dbRO.Health())
	if err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(jsonResp); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) headerTextsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Header Texts Handler]"
	settings, err := requests.GetSettingsData(
		r.Context(),
		s.cache,
		&s.SF,
		s.dbRO,
		[]byte("key_header_texts"),
		[]string{"email", "mobile_no"},
	)
	if err != nil {
		logs.Log().Error(logtag, zap.Error(err))
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
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) footerTextsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Footer Texts Handler]"
	settings, err := requests.GetSettingsData(
		r.Context(),
		s.cache,
		&s.SF,
		s.dbRO,
		[]byte("key_footer_texts"),
		[]string{"mobile_no", "email", "url_gmap", "url_facebook", "url_tiktok"},
	)
	if err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	texts := []models.FooterRowText{
		{
			Label: "Home",
			URL:   "/cchoice/",
		},
		{
			Label: "About Us",
			URL:   "/cchoice#about-us",
		},
		{
			Label: "Services",
			URL:   "/cchoice#services",
		},
		{
			Label: "Partners",
			URL:   "/cchoice#partners",
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
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Search Handler]"
	if err := r.ParseForm(); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	search := r.PostFormValue("search")
	products, err := s.dbRO.GetQueries().GetProductsBySearchQuery(
		r.Context(),
		queries.GetProductsBySearchQueryParams{
			Name:  search,
			Limit: constants.MaxSearchShowResults,
		},
	)
	if err != nil || len(products) == 0 {
		logs.Log().Info(logtag, zap.String("query", search))
		return
	}

	logs.Log().Info(
		logtag,
		zap.Int("count", len(products)),
		zap.Int("limit", constants.MaxSearchShowResults),
		zap.String("query", search),
	)

	for _, product := range products {
		if strings.HasSuffix(product.ThumbnailPath, constants.EmptyImageFilename) {
			continue
		}

		finalPath, ext, err := images.GetImagePathWithSize(product.ThumbnailPath, constants.DefaultThumbnailSize, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			continue
		}

		imgData, err := images.GetImageDataB64(s.cache, s.fs, finalPath, ext)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			continue
		}

		product.ThumbnailData = imgData

		if err := components.SearchResultProductCard(models.ToSearchResultProduct(s.encoder, product)).Render(r.Context(), w); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			return
		}
	}

	if err := components.SearchMore(search).Render(r.Context(), w); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		return
	}
}

func (s *Server) checkoutsHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Checkouts Handler]"
	if err := r.ParseForm(); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch s.paymentGateway.GatewayEnum() {
	case payments.PAYMENT_GATEWAY_PAYMONGO:
		// if err := s.paymentGateway.CheckoutPaymentHandler(w, r); err != nil {
		// 	logs.Log().Error("[PayMongo] Checkouts handler", zap.Error(err))
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
	default:
		err := errors.New("checkouts handler. Unimplemented payment gateway")
		logs.Log().Error(err.Error(), zap.String("gateway", s.paymentGateway.GatewayEnum().String()))
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}
}
