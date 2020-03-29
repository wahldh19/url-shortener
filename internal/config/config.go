package config

import "os"

const (
	defaultEnvironment = "development"
	defaultMongoURI    = "mongodb://localhost:27017/?readConcernLevel=majority&retryWrites=true&w=majority"
	defaultPort        = "8080"
)

// Config represents a service configuration.
type Config struct {
	// Environment is the service deployment environment.
	Environment string

	// MongoURI is a MongoDB URI connection string.
	MongoURI string

	// Port is the port used to listen for HTTP requests.
	Port string
}

// New returns a service Config.
// It will attempt to get and use the following environment variables:
// ENV
// MONGO_URI
// PORT
// If these variables are not set, it will default to the constants
// defined in this package.
func New() Config {
	return Config{
		Environment: func() string {
			if env := os.Getenv("ENV"); env != "" {
				return env
			}
			return defaultEnvironment
		}(),
		MongoURI: func() string {
			if mongoURI := os.Getenv("MONGO_URI"); mongoURI != "" {
				return mongoURI
			}
			return defaultMongoURI
		}(),
		Port: func() string {
			if port := os.Getenv("PORT"); port != "" {
				return port
			}
			return defaultPort
		}(),
	}
}
