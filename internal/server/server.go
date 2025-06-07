package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/sync/singleflight"

	"cchoice/internal/database"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
)

const CACHE_MAX_BYTES int = 1024

type Server struct {
	dbRO      database.Service
	dbRW      database.Service
	SF        singleflight.Group
	fs        http.FileSystem
	fsHandler http.Handler
	fsServer  *http.Server
	Cache     *fastcache.Cache
	address   string
	port      int
	portFS    int
	secure    bool
	useHTTP2  bool
	useSSL    bool
}

func NewServer() *http.Server {
	address := os.Getenv("ADDRESS")
	if address == "" {
		panic("No ADDRESS set")
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}

	portFS, err := strconv.Atoi(os.Getenv("PORT_FS"))
	if err != nil {
		panic(err)
	}

	dbRO := database.New(database.DB_MODE_RO)
	dbRW := database.New(database.DB_MODE_RW)
	NewServer := &Server{
		address:  address,
		port:     port,
		portFS:   portFS,
		dbRO:     dbRO,
		dbRW:     dbRW,
		Cache:    fastcache.New(CACHE_MAX_BYTES),
		useHTTP2: utils.GetBoolFlag(os.Getenv("USEHTTP2")),
		useSSL:   utils.GetBoolFlag(os.Getenv("USESSL")),
	}

	addr := fmt.Sprintf("%s:%d", NewServer.address, NewServer.port)
	readTimeout := 10 * time.Second
	writeTimeout := 30 * time.Second

	var tlsConfig *tls.Config
	if NewServer.useSSL {
		var certPath, keyPath string
		if os.Getenv("APP_ENV") == "local" {
			certPath = os.Getenv("CERTPATH")
			keyPath = os.Getenv("KEYPATH")
		} else {
			certPath = fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", NewServer.address)
			keyPath = fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", NewServer.address)
		}
		logs.Log().Info(
			"SSL: opening files",
			zap.String("cert", certPath),
			zap.String("key", keyPath),
		)
		serverTLSCert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			panic(err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{serverTLSCert},
		}
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      NewServer.RegisterRoutes(),
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
		zap.String("address", addr),
		zap.Bool("SSL", NewServer.useSSL),
		zap.Bool("HTTP2", NewServer.useHTTP2),
		zap.Duration("read timeout", readTimeout),
		zap.Duration("write timeout", writeTimeout),
	)

	return server
}
