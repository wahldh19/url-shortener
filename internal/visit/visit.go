package visit

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// Context timeouts for DB operations.
	aggregateTimeout = 2 * time.Second
	insertTimeout    = 1 * time.Second

	// collection is the MongoDB collection for Visit documents.
	Collection = "visits"
)

// Visit represents a document storing a visit to a short URL.
type Visit struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	ShortID primitive.ObjectID `bson:"shortId"`
	Time    time.Time          `bson:"time"`
}

// Stats represent counts of visits to short URLs.
type Stats struct {
	Day  int `bson:"day" json:"day"`
	Week int `bson:"week" json:"week"`
	Year int `bson:"year" json:"year"`
}
