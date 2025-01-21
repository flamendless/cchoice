package server

import (
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
)

const CACHE_MAX_BYTES int = 1024

type Server struct {
	address string
	port    int
	dbRO    database.Service
	dbRW    database.Service
	Cache   *fastcache.Cache
	SF      singleflight.Group
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

	logs.Log().Info(
		"Server",
		zap.String("address", addr),
		zap.Duration("read timeout", readTimeout),
		zap.Duration("write timeout", writeTimeout),
	)

	h2s := &http2.Server{MaxConcurrentStreams: 256}
	h2cHandler := h2c.NewHandler(NewServer.RegisterRoutes(), h2s)
	server := &http.Server{
		Addr:         addr,
		Handler:      h2cHandler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return server
}
