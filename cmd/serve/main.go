package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dwrz/url-shortener/internal/config"
	"github.com/dwrz/url-shortener/internal/db"
)

func main() {

	log.Println("starting")

	// Get the service configuration.
	cfg := config.New()

	// Setup the main context.
	ctx, cancel := context.WithCancel(context.Background())

	// Connect to the DB.
	db, err := db.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}

	// Start and run the HTTP server.
	serverDone := make(chan struct{})
	go serve(ctx, serveParams{
		db:          db,
		done:        serverDone,
		environment: cfg.Environment,
		port:        cfg.Port,
	})

	// Listen for OS signals.
	osListener := make(chan os.Signal, 1)
	signal.Notify(
		osListener,
		syscall.SIGTERM, syscall.SIGINT,
	)

	// If we receive a signal, cancel the main context.
	s := <-osListener
	log.Printf("received signal: %s", s)
	cancel()

	// Wait for the server to report shutdown.
	<-serverDone

	log.Println("terminating")
}
