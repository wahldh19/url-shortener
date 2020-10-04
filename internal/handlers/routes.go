package handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

const (
	// pathCreate is the endpoint used to create a new short URL.
	pathCreate = "/"

	// pathRedirect is the endpoint used to redirect a short URL.
	pathRedirect = "/{short}"

	// pathStats is the endpoint used to get stats for a short URL.
	pathStats = "/{short}/stats"

	// pathStatus is the healthcheck endpoint for the service.
	pathStatus = "/"
)

type AddRoutesParams struct {
	// DB client handlers should inject for MongoDB queries.
	DB *mongo.Client

	// Environment the handlers should function in.
	// This value should be taken from the service configuration.
	Environment string

	// Router which routes should be added to.
	Router *mux.Router
}

func (p *AddRoutesParams) validate() error {
	if p.DB == nil {
		return fmt.Errorf("missing db client")
	}
	if p.Environment == "" {
		return fmt.Errorf("missing environment")
	}
	if p.Router == nil {
		return fmt.Errorf("missing router")
	}

	return nil
}

// AddRoutes attaches handlers to the Router, and sets the DB and
// Environment configuration on handlers.
func AddRoutes(p AddRoutesParams) error {
	if err := p.validate(); err != nil {
		return fmt.Errorf("invalid params: %v", err)
	}

	// Add the status handler.
	p.Router.HandleFunc(
		pathStatus,
		handler{db: p.DB, environment: p.Environment}.Status,
	).Methods(http.MethodGet)

	// Add the create handler.
	p.Router.HandleFunc(
		pathCreate,
		handler{db: p.DB, environment: p.Environment}.Create,
	).Methods(http.MethodPost)

	// Add the redirect handler.
	p.Router.HandleFunc(
		pathRedirect,
		handler{db: p.DB, environment: p.Environment}.Redirect,
	).Methods(http.MethodGet)

	// Add the stats handler.
	p.Router.HandleFunc(
		pathStats,
		handler{db: p.DB, environment: p.Environment}.Stats,
	).Methods(http.MethodGet)

	return nil
}
