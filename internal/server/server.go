package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"cchoice/internal/database"
)

type Server struct {
	address string
	port    int
	dbRO    database.Service
	dbRW    database.Service
}

func NewServer() *http.Server {
	address := os.Getenv("ADDRESS")
	if address != "" {
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

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", NewServer.address, NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
