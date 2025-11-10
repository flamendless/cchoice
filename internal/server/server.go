package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
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
	"cchoice/internal/geocoding/googlemaps"
	"cchoice/internal/logs"
	"cchoice/internal/payments"
	"cchoice/internal/payments/paymongo"
	"cchoice/internal/shipping"
	cchoiceservice "cchoice/internal/shipping/cchoice"
	"cchoice/internal/shipping/lalamove"
	"cchoice/internal/storage"
	"cchoice/internal/storage/linode"
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
	objectStorage   storage.IObjectStorage
	geocoder        geocoding.IGeocoder
	encoder         encode.IEncode
	address         string
	port            int
	portFS          int
	useHTTP2        bool
	useSSL          bool
}

func (s *Server) buildURL(path string) string {
	//TODO: (Brandon) Implement CDN
	return fmt.Sprintf(
		`https://%s/cchoice/%s`,
		s.address,
		path,
	)
}

func NewServer() *http.Server {
	cfg := conf.Conf()
	sessionManager := scs.New()
	sessionManager.Lifetime = 1 * time.Hour

	dbRO := database.New(database.DB_MODE_RO)
	dbRW := database.New(database.DB_MODE_RW)

	var paymentGateway payments.IPaymentGateway
	switch cfg.PaymentService {
	case "paymongo":
		paymentGateway = paymongo.MustInit()
	default:
		panic("Unsupported payment service: " + cfg.PaymentService)
	}

	var shippingService shipping.IShippingService
	switch cfg.ShippingService {
	case "lalamove":
		shippingService = lalamove.MustInit()
	case "cchoice":
		shippingService = cchoiceservice.MustInit()
	default:
		panic("Unsupported shipping service: " + cfg.ShippingService)
	}

	var geocoder geocoding.IGeocoder
	switch cfg.GeocodingService {
	case "googlemaps":
		geocoder = googlemaps.MustInit(dbRW)
	default:
		panic("Unsupported geocoding service: " + cfg.GeocodingService)
	}

	staticFS := localstorage.New()

	var objStorage storage.IObjectStorage
	var productImageFS storage.IFileSystem

	switch cfg.StorageProvider {
	case "linode":
		objStorage = linode.MustInit()
		productImageFS = linode.New(objStorage)
	case "local":
		objStorage = nil
		productImageFS = localstorage.New()
	default:
		panic("Unsupported storage provider: " + cfg.StorageProvider)
	}

	NewServer := &Server{
		address:         cfg.Server.Address,
		port:            cfg.Server.Port,
		portFS:          cfg.Server.PortFS,
		dbRO:            dbRO,
		dbRW:            dbRW,
		staticFS:        staticFS,
		productImageFS:  productImageFS,
		cache:           fastcache.New(CACHE_MAX_BYTES),
		sessionManager:  sessionManager,
		paymentGateway:  paymentGateway,
		shippingService: shippingService,
		objectStorage:   objStorage,
		geocoder:        geocoder,
		encoder:         sqids.MustSqids(),
		useHTTP2:        cfg.Server.UseHTTP2,
		useSSL:          cfg.Server.UseSSL,
	}

	addr := fmt.Sprintf("%s:%d", NewServer.address, NewServer.port)
	readTimeout := 10 * time.Second
	writeTimeout := 30 * time.Second

	var tlsConfig *tls.Config
	if NewServer.useSSL {
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

	handler := sessionManager.LoadAndSave(NewServer.RegisterRoutes())

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		TLSConfig:    tlsConfig,
	}
	if NewServer.useHTTP2 {
		if err := http2.ConfigureServer(server, &http2.Server{
			MaxConcurrentStreams: 256,
		}); err != nil {
			logs.Log().Error("Server configure", zap.Error(err))
		}
	}

	logs.Log().Info(
		"Server",
		zap.String("Address", addr),
		zap.Bool("Use caching", NewServer.cache != nil),
		zap.Int("Caching max bytes", CACHE_MAX_BYTES),
		zap.Bool("Use session manager", NewServer.sessionManager != nil),
		zap.Duration("Session manager lifetime", NewServer.sessionManager.Lifetime),
		zap.String("Payment gateway", NewServer.paymentGateway.GatewayEnum().String()),
		zap.String("Shipping service", NewServer.shippingService.Enum().String()),
		zap.String("Storage provider", cfg.StorageProvider),
		zap.Bool("SSL", NewServer.useSSL),
		zap.Bool("HTTP2", NewServer.useHTTP2),
		zap.Duration("Read timeout", readTimeout),
		zap.Duration("Write timeout", writeTimeout),
	)

	return server
}
