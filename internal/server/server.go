package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/alexedwards/scs/v2"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/sync/singleflight"

	"cchoice/internal/conf"
	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/encode"
	"cchoice/internal/encode/sqids"
	"cchoice/internal/enums"
	"cchoice/internal/geocoding"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"
	"cchoice/internal/mail"
	"cchoice/internal/payments"
	"cchoice/internal/requests"
	"cchoice/internal/services"
	"cchoice/internal/shipping"
	"cchoice/internal/storage"
	localstorage "cchoice/internal/storage/local"
	"cchoice/internal/utils"
)

type Services struct {
	product      *services.ProductService
	productImage *services.ProductImageService
	brand        *services.BrandService
	staff        *services.StaffService
	staffLog     *services.StaffLogsService
	role         *services.RoleService
	location     *services.LocationService
	attendance   *services.AttendanceService
	report       *services.ReportService
	customer     *services.CustomerService
}

type Server struct {
	dbRO               database.IService
	dbRW               database.IService
	SF                 singleflight.Group
	staticFS           http.FileSystem // For static assets (JS, CSS, icons) - always local
	productImageFS     http.FileSystem // For product images - configurable (local or object storage)
	cache              *fastcache.Cache
	sessionManager     *scs.SessionManager
	paymentGateway     payments.IPaymentGateway
	shippingService    shipping.IShippingService
	geocoder           geocoding.IGeocoder
	objectStorage      storage.IObjectStorage
	encoder            encode.IEncode
	mailService        mail.IMailService
	mailJobRunner      *jobs.EmailJobRunner
	thumbnailService   jobs.IThumbnailService
	thumbnailJobRunner *jobs.ThumbnailJobRunner
	services           Services
	address            string
	port               int
	portFS             int
	useHTTP2           bool
	useSSL             bool
}

func (s *Server) GetProductImageProxyURL(ctx context.Context, thumbnailPath string, size string) (string, error) {
	pathToUse := thumbnailPath
	if size != "" {
		parts := strings.Split(pathToUse, "/")
		for i, part := range parts {
			if part == "webp" && i+1 < len(parts) {
				parts[i+1] = size
				pathToUse = strings.Join(parts, "/")
				break
			}
		}
	}

	cfg := conf.Conf()
	if cfg.IsLocal() && s.objectStorage != nil {
		presignedURL, err := s.objectStorage.PresignedGetObject(ctx, pathToUse, 24*time.Hour)
		if err != nil {
			return "", fmt.Errorf("failed to generate presigned URL: %w", err)
		}
		return presignedURL, nil
	}

	proxyURL := fmt.Sprintf("https://%s%s?path=%s&thumbnail=1&quality=best", s.address, utils.URL("/products/image"), url.QueryEscape(pathToUse))
	if size != "" {
		proxyURL += "&size=" + url.QueryEscape(size)
	}
	return proxyURL, nil
}

func NewServer() *ServerInstance {
	cfg := conf.Conf()
	sessionManager := scs.New()
	sessionManager.Lifetime = 1 * time.Hour

	dbRO := database.New(database.DB_MODE_RO)
	dbRW := database.New(database.DB_MODE_RW)

	var objStorage storage.IObjectStorage
	var productImageFS http.FileSystem
	var mailService mail.IMailService
	var emailJobRunner *jobs.EmailJobRunner
	var paymentGateway payments.IPaymentGateway
	var shippingService shipping.IShippingService
	var geocoder geocoding.IGeocoder

	if cfg.IsWeb() {
		objStorage, productImageFS = mustInitStorageProvider()
		logs.Log().Info("Web mode: skipping payment, shipping, geocoding, mail services")
	} else {
		objStorage, productImageFS = mustInitStorageProvider()
		mailService = mustInitMailService()
		emailJobRunner = jobs.NewEmailJobRunner(dbRW.GetDB(), dbRO, dbRW, mailService)
		paymentGateway = mustInitPaymentGateway()
		shippingService = mustInitShippingService()
		geocoder = mustInitGeocodingService(dbRW)
	}

	var thumbnailService *services.ThumbnailService
	var thumbnailJobRunner *jobs.ThumbnailJobRunner
	thumbnailService = services.NewThumbnailService(objStorage)
	thumbnailJobRunner = jobs.NewThumbnailJobRunner(dbRW.GetDB(), dbRO, dbRW, thumbnailService)

	newServer := &Server{
		address:            cfg.Server.Address,
		port:               cfg.Server.Port,
		portFS:             cfg.Server.PortFS,
		dbRO:               dbRO,
		dbRW:               dbRW,
		staticFS:           localstorage.New(),
		productImageFS:     productImageFS,
		cache:              fastcache.New(constants.CacheMaxBytes),
		sessionManager:     sessionManager,
		paymentGateway:     paymentGateway,
		shippingService:    shippingService,
		objectStorage:      objStorage,
		geocoder:           geocoder,
		encoder:            sqids.MustSqids(),
		mailService:        mailService,
		mailJobRunner:      emailJobRunner,
		thumbnailService:   thumbnailService,
		thumbnailJobRunner: thumbnailJobRunner,
		useHTTP2:           cfg.Server.UseHTTP2,
		useSSL:             cfg.Server.UseSSL,
	}

	staffLogService := services.NewStaffLogsService(newServer.encoder, newServer.dbRO, newServer.dbRW)
	locationService := services.NewLocationService(cfg.Settings.ShopLocation)
	attendanceService := services.NewAttendanceService(newServer.encoder, newServer.dbRO, newServer.dbRW)

	newServer.services = Services{
		product:      services.NewProductService(newServer.encoder, newServer.dbRO, newServer.dbRW, newServer.GetCDNURL),
		productImage: services.NewProductImageService(newServer.objectStorage, newServer.encoder, newServer.dbRO, newServer.dbRW),
		brand:        services.NewBrandService(newServer.encoder, newServer.dbRO, newServer.dbRW),
		staff:        services.NewStaffServiceWithDeps(newServer.encoder, newServer.dbRO, newServer.dbRW, attendanceService, locationService),
		staffLog:     staffLogService,
		role:         services.NewRoleService(newServer.encoder, newServer.dbRO, newServer.dbRW),
		location:     locationService,
		attendance:   attendanceService,
		report:       services.NewReportService(newServer.encoder, newServer.dbRO, staffLogService),
		customer:     services.NewCustomerService(newServer.encoder, newServer.dbRO, newServer.dbRW),
	}

	ctx := context.Background()
	settings, err := requests.GetSettingsData(
		ctx,
		newServer.cache,
		&newServer.SF,
		newServer.dbRO,
		[]byte("server_settings"),
		[]string{
			"mobile_no",
			"email",
			"address",
			"url_gmap",
			"url_waze",
			"url_facebook",
			"url_tiktok",
			"show_promo_banner",
			"shop_location",
			"vat_percentage",
		},
	)
	if err != nil {
		logs.LogCtx(ctx).Error("[New Server] failed to get settings", zap.Error(err))
		return nil
	}
	cfg.SetSettings(settings)

	var addr string
	switch cfg.AppEnv {
	case enums.APP_ENV_LOCAL, enums.APP_ENV_WEB:
		addr = fmt.Sprintf("%s:%d", newServer.address, newServer.port)
	case enums.APP_ENV_PROD:
		addr = fmt.Sprintf(":%d", newServer.port)
	}

	var tlsConfig *tls.Config
	if newServer.useSSL {
		logs.Log().Info(
			"SSL: opening files",
			zap.String("cert", cfg.Server.CertPath),
			zap.String("key", cfg.Server.KeyPath),
		)
		serverTLSCert, err := tls.LoadX509KeyPair(cfg.Server.CertPath, cfg.Server.KeyPath)
		if err != nil {
			panic(err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{serverTLSCert},
		}
	}

	readTimeout := 10 * time.Second
	writeTimeout := 30 * time.Second

	handler := sessionManager.LoadAndSave(newServer.RegisterRoutes())

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		TLSConfig:    tlsConfig,
	}
	if newServer.useHTTP2 {
		if err := http2.ConfigureServer(httpServer, &http2.Server{
			MaxConcurrentStreams: 256,
		}); err != nil {
			logs.Log().Error("Server configure", zap.Error(err))
		}
	}

	logFields := []zap.Field{
		zap.String("Address", addr),
		zap.Stringer("AppEnv", cfg.AppEnv),
		zap.Bool("Use caching", newServer.cache != nil),
		zap.Int("Caching max bytes", constants.CacheMaxBytes),
		zap.Bool("Use session manager", newServer.sessionManager != nil),
		zap.Duration("Session manager lifetime", newServer.sessionManager.Lifetime),
		zap.Bool("SSL", newServer.useSSL),
		zap.Bool("HTTP2", newServer.useHTTP2),
		zap.Duration("Read timeout", readTimeout),
		zap.Duration("Write timeout", writeTimeout),
		zap.String("Encoder", newServer.encoder.Name()),
		zap.Any("Tests", cfg.Test),
	}

	if newServer.shippingService != nil {
		logFields = append(logFields, zap.String("Shipping service", newServer.shippingService.Enum().String()))
	}
	if newServer.geocoder != nil {
		logFields = append(logFields, zap.String("Geocoder service", newServer.geocoder.Enum().String()))
	}
	if newServer.objectStorage != nil {
		logFields = append(logFields, zap.String("Storage provider", newServer.objectStorage.ProviderEnum().String()))
	}
	if newServer.mailService != nil {
		logFields = append(logFields, zap.String("Mail service", newServer.mailService.Enum().String()))
	}

	logs.Log().Info("Server Config", logFields...)

	return &ServerInstance{
		HTTPServer: httpServer,
		internal:   newServer,
	}
}
