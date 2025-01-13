package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	"cchoice/internal/database"
	"cchoice/internal/logs"
)

type Server struct {
	address string
	port    int
	dbRO    database.Service
	dbRW    database.Service
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

	NewServer := &Server{
		address: address,
		port:    port,
		dbRO:    database.New(database.DB_MODE_RO),
		dbRW:    database.New(database.DB_MODE_RW),
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

	server := &http.Server{
		Addr:         addr,
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return server
}
