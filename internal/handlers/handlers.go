package handlers

import (
	"encoding/json"
	"github.com/dwrz/url-shortener/internal/shorturl"
	"github.com/dwrz/url-shortener/internal/validurl"
	"github.com/dwrz/url-shortener/internal/visit"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
)

// handler is used to store values needed by methods implementing the
// net/http Handler interface for this service.
type handler struct {
	// db is the client handlers should inject for MongoDB queries.
	db *mongo.Client

	// environment is the service environment handlers should operate
	// in; e.g., "development", "staging", or "production".
	// This value determines which database to use for MongoDB.
	// It should originate from the service configuration.
	environment string
}

// Create handlers requests to create a new short URL.
// It expects a long URL in an application/x-www-form-urlencoded body,
// with the field for the long URL as "url".
// It returns the short code for the URL in a response body.
func (h handler) Create(w http.ResponseWriter, r *http.Request) {
	longURL := r.FormValue("url")

	// Validate the URL.
	if err := validurl.Validate(longURL); err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	// Generate the short URL and create a DB document.
	short, err := shorturl.Create(r.Context(), shorturl.CreateParams{
		DB:          h.db,
		Environment: h.environment,
		LongURL:     longURL,
	})
	if err != nil {
		log.Printf("failed to create short url: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// Respond with the short URL id.
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(short))
}

// Redirect gets the requested short URL id, and if it exists, redirects
// the client to the associated long URL. It also stores a record of the
// visit to this short URL.
func (h handler) Redirect(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)

	// Retrieve the short URL.
	s, err := shorturl.Get(r.Context(), shorturl.GetParams{
		DB:          h.db,
		Environment: h.environment,
		Short:       pathParams["short"],
	})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Create a visit record.
	if err := visit.Create(r.Context(), visit.CreateParams{
		DB:          h.db,
		Environment: h.environment,
		ShortID:     s.ID,
	}); err != nil {
		log.Printf("failed to create visit record: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// Redirect to the associated long URL.
	http.Redirect(w, r, s.URL, http.StatusMovedPermanently)
}

// Stats gets the visit count for a short URL.
// It responds with a application/json body which specifies the number
// of visits in the past 24 hours, week, and year.
func (h handler) Stats(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)

	// Retrieve the short URL.
	s, err := shorturl.Get(r.Context(), shorturl.GetParams{
		DB:          h.db,
		Environment: h.environment,
		Short:       pathParams["short"],
	})
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Get the visited stats for this short URL.
	stats, err := visit.GetStats(r.Context(), visit.GetStatsParams{
		DB:          h.db,
		Environment: h.environment,
		ShortID:     s.ID,
	})
	if err != nil {
		log.Printf("failed to get stats: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// Respond with JSON encoded stats.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("failed to json encode stats: %v", err)
	}

}

// status is used as a health check for this service.
// It should respond with a 200 HTTP Status OK and an empty body.
func (h handler) Status(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Query().Get("stats") == "true":
		shortURLCount, err := shorturl.CountURLs(
			r.Context(), shorturl.CountParams{
				DB:          h.db,
				Environment: h.environment,
			},
		)
		if err != nil {
			http.Error(
				w, err.Error(), http.StatusInternalServerError,
			)
			return
		}
		response := struct {
			ShortURLCount int64 `json:"shortURLCount"`
		}{ShortURLCount: shortURLCount}
		j, err := json.Marshal(response)
		if err != nil {
			http.Error(
				w, err.Error(), http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	default:
		w.Write([]byte{})
	}
}
