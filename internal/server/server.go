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
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/singleflight"

	"cchoice/internal/database"
	"cchoice/internal/logs"
	"cchoice/internal/utils"
)

const CACHE_MAX_BYTES int = 1024

type Server struct {
	dbRO    database.Service
	dbRW    database.Service
	SF      singleflight.Group
	fs      http.FileSystem
	Cache   *fastcache.Cache
	address string
	port    int
	secure  bool
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

	dbRO := database.New(database.DB_MODE_RO)
	dbRW := database.New(database.DB_MODE_RW)
	NewServer := &Server{
		address: address,
		port:    port,
		dbRO:    dbRO,
		dbRW:    dbRW,
		Cache:   fastcache.New(CACHE_MAX_BYTES),
	}

	addr := fmt.Sprintf("%s:%d", NewServer.address, NewServer.port)
	readTimeout := 10 * time.Second
	writeTimeout := 30 * time.Second

	var tlsConfig *tls.Config
	useSSL := utils.GetBoolFlag(os.Getenv("USESSL"))
	if useSSL {
		certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", NewServer.address)
		keyPath := fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", NewServer.address)
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

	useHTTP2 := utils.GetBoolFlag(os.Getenv("USEHTTP2"))
	logs.Log().Info(
		"Server",
		zap.String("address", addr),
		zap.Bool("SSL", useSSL),
		zap.Bool("HTTP2", useHTTP2),
		zap.Duration("read timeout", readTimeout),
		zap.Duration("write timeout", writeTimeout),
	)

	var protocol *http.Protocols
	handler := NewServer.RegisterRoutes()
	if useHTTP2 {
		h2s := &http2.Server{MaxConcurrentStreams: 256}
		handler = h2c.NewHandler(handler, h2s)

		t := http.DefaultTransport.(*http.Transport).Clone()
		t.Protocols = new(http.Protocols)
		t.Protocols.SetHTTP1(true)
		t.Protocols.SetHTTP2(true)
		protocol = t.Protocols
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		TLSConfig:    tlsConfig,
		Protocols:    protocol,
	}
	return server
}
