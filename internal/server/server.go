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
	"cchoice/internal/database"
	"cchoice/internal/encode"
	"cchoice/internal/encode/sqids"
	"cchoice/internal/geocoding"
	"cchoice/internal/jobs"
	"cchoice/internal/logs"
	"cchoice/internal/mail"
	"cchoice/internal/payments"
	"cchoice/internal/shipping"
	"cchoice/internal/storage"
	localstorage "cchoice/internal/storage/local"
)

const CACHE_MAX_BYTES int = 100 * 1024 * 1024 // 100MB cache for better cost efficiency

type Server struct {
	dbRO            database.Service
	dbRW            database.Service
	SF              singleflight.Group
	staticFS        http.FileSystem // For static assets (JS, CSS, icons) - always local
	productImageFS  http.FileSystem // For product images - configurable (local or object storage)
	cache           *fastcache.Cache
	sessionManager  *scs.SessionManager
	paymentGateway  payments.IPaymentGateway
	shippingService shipping.IShippingService
	geocoder        geocoding.IGeocoder
	objectStorage   storage.IObjectStorage
	encoder         encode.IEncode
	mailService     mail.IMailService
	emailJobRunner  *jobs.EmailJobRunner
	address         string
	port            int
	portFS          int
	useHTTP2        bool
	useSSL          bool
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

	proxyURL := fmt.Sprintf("https://%s/cchoice/products/image?path=%s&thumbnail=1&quality=best", s.address, url.QueryEscape(pathToUse))
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
	objStorage, productImageFS := mustInitStorageProvider()
	mailService := mustInitMailService()

	emailJobRunner := jobs.NewEmailJobRunner(dbRW.GetDB(), dbRO, dbRW, mailService)

	newServer := &Server{
		address:         cfg.Server.Address,
		port:            cfg.Server.Port,
		portFS:          cfg.Server.PortFS,
		dbRO:            dbRO,
		dbRW:            dbRW,
		staticFS:        localstorage.New(),
		productImageFS:  productImageFS,
		cache:           fastcache.New(CACHE_MAX_BYTES),
		sessionManager:  sessionManager,
		paymentGateway:  mustInitPaymentGateway(),
		shippingService: mustInitShippingService(),
		objectStorage:   objStorage,
		geocoder:        mustInitGeocodingService(dbRW),
		encoder:         sqids.MustSqids(),
		mailService:     mailService,
		emailJobRunner:  emailJobRunner,
		useHTTP2:        cfg.Server.UseHTTP2,
		useSSL:          cfg.Server.UseSSL,
	}

	var addr string
	switch cfg.AppEnv {
	case "local":
		addr = fmt.Sprintf("%s:%d", newServer.address, newServer.port)
	case "prod":
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

	logs.Log().Info(
		"Server Config",
		zap.String("Address", addr),
		zap.Bool("Use caching", newServer.cache != nil),
		zap.Int("Caching max bytes", CACHE_MAX_BYTES),
		zap.Bool("Use session manager", newServer.sessionManager != nil),
		zap.Duration("Session manager lifetime", newServer.sessionManager.Lifetime),
		zap.Bool("SSL", newServer.useSSL),
		zap.Bool("HTTP2", newServer.useHTTP2),
		zap.Duration("Read timeout", readTimeout),
		zap.Duration("Write timeout", writeTimeout),
		zap.String("Payment gateway", newServer.paymentGateway.GatewayEnum().String()),
		zap.String("Shipping service", newServer.shippingService.Enum().String()),
		zap.String("Geocoder service", newServer.geocoder.Enum().String()),
		zap.String("Storage provider", newServer.objectStorage.ProviderEnum().String()),
		zap.String("Encoder", newServer.encoder.Name()),
		zap.String("Mail service", newServer.mailService.Enum().String()),
	)

	return &ServerInstance{
		HTTPServer: httpServer,
		internal:   newServer,
	}
}
