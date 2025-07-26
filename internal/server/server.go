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
	"cchoice/internal/logs"
	"cchoice/internal/payments"
	"cchoice/internal/payments/paymongo"
)

const CACHE_MAX_BYTES int = 1024

type Server struct {
	dbRO           database.Service
	dbRW           database.Service
	SF             singleflight.Group
	fs             http.FileSystem
	cache          *fastcache.Cache
	sessionManager *scs.SessionManager
	paymentGateway payments.IPaymentGateway
	encoder        encode.IEncode
	address        string
	port           int
	portFS         int
	useHTTP2       bool
	useSSL         bool
}

func NewServer() *http.Server {
	cfg := conf.Conf()
	sessionManager := scs.New()
	sessionManager.Lifetime = 1 * time.Hour

	dbRO := database.New(database.DB_MODE_RO)
	dbRW := database.New(database.DB_MODE_RW)
	NewServer := &Server{
		address:        cfg.Address,
		port:           cfg.Port,
		portFS:         cfg.PortFS,
		dbRO:           dbRO,
		dbRW:           dbRW,
		cache:          fastcache.New(CACHE_MAX_BYTES),
		sessionManager: sessionManager,
		paymentGateway: paymongo.MustInit(),
		encoder:        sqids.MustSqids(),
		useHTTP2:       cfg.UseHTTP2,
		useSSL:         cfg.UseSSL,
	}

	addr := fmt.Sprintf("%s:%d", NewServer.address, NewServer.port)
	readTimeout := 10 * time.Second
	writeTimeout := 30 * time.Second

	var tlsConfig *tls.Config
	if NewServer.useSSL {
		logs.Log().Info(
			"SSL: opening files",
			zap.String("cert", cfg.CertPath),
			zap.String("key", cfg.KeyPath),
		)
		serverTLSCert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
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
		zap.Bool("SSL", NewServer.useSSL),
		zap.Bool("HTTP2", NewServer.useHTTP2),
		zap.Duration("Read timeout", readTimeout),
		zap.Duration("Write timeout", writeTimeout),
	)

	return server
}
