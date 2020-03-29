package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dwrz/url-shortener/internal/handlers"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

// shutdownTimeout is how long the server will wait to process existing
// requests before shutting down. The value should be set in accordance
// with the hosting environment, or refactored to be set via config.
const shutdownTimeout = 30 * time.Second

type serveParams struct {
	db          *mongo.Client
	done        chan struct{}
	environment string
	port        string
}

func (p serveParams) validate() error {
	if p.db == nil {
		return fmt.Errorf("missing db client")
	}
	if p.done == nil {
		return fmt.Errorf("missing done channel")
	}
	if p.environment == "" {
		return fmt.Errorf("missing environment")
	}
	if p.port == "" {
		return fmt.Errorf("missing port")
	}

	return nil
}

func serve(ctx context.Context, p serveParams) {
	if err := p.validate(); err != nil {
		log.Fatalf("invalid server configuration parameters: %v", err)
	}

	// Create the HTTP router and attach the handlers.
	router := mux.NewRouter()

	if err := handlers.AddRoutes(handlers.AddRoutesParams{
		DB:          p.db,
		Environment: p.environment,
		Router:      router,
	}); err != nil {
		log.Fatalf("failed to add handlers to mux router: %v", err)
	}

	// Setup and start the server.
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", p.port),
		Handler: router,
	}

	go func() {
		log.Println("starting http server")
		log.Printf("using port %s", p.port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("failed to listen and serve: %v", err)
		}
	}()

	// Block until the main context is canceled.
	<-ctx.Done()

	// Shutdown the HTTP server.
	log.Println("stopping http server")
	ctxShutdown, cancel := context.WithTimeout(
		context.Background(), shutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("failed to shutdown http server: %v", err)
	}
	log.Println("http server shutdown")

	// Signal to the main goroutine that shutdown is complete.
	close(p.done)
}
