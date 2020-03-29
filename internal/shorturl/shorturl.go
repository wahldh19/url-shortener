package shorturl

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// Context timeouts for DB operations.
	findTimeout   = 1 * time.Second
	insertTimeout = 1 * time.Second

	// Collection is the MongoDB collection for ShortURL documents.
	Collection = "urls"

	// defaultLength is the starting length of short URL strings.
	// New short URLs will be generated with this length.
	defaultLength = 6

	// maxLength is the maximum length of short URL strings.
	// If a unique short URL id cannot be generated with a string of
	// this length, an error is returned.
	maxLength = 8
)

// ShortURL represents a document storing a short URL and its
// corresponding long URL.
type ShortURL struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Created time.Time          `bson:"created"`

	// Short is the short URL string used to retrieve the long URL.
	Short string `bson:"short"`

	// URL is the original long URL for which a short URL was
	// created.
	URL string `bson:"url"`
}
