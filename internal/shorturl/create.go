package shorturl

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dwrz/url-shortener/pkg/randstr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateParams struct {
	DB          *mongo.Client
	Environment string
	LongURL     string
}

func (p CreateParams) validate() error {
	if p.DB == nil {
		return fmt.Errorf("missing db client")
	}
	if p.Environment == "" {
		return fmt.Errorf("invalid env")
	}
	if p.LongURL == "" {
		return fmt.Errorf("missing long url")
	}

	return nil
}

// Create generates a short URL string for the input long URL, and
// inserts a new ShortURL document into MongoDB.
// Create attempts to create a short URL string of defaultLength.
// If there is a collision, a new string is generated with an
// incremented length -- which should reduce the likelihood of another
// collision. This is repeated until maxLength is reached, at which
// point Create errors out.
// NB: Because of the nature of concurrent distributed systems, there
// still exists a small possibility of a collision here, with a write
// concurrent to a stale read. This possibility is so small as to be
// practically worth ignoring, but it does exit.
// Increasing the defaultLength of the short URL decreases this
// possibility, as would using a read concern of linearizable.
// The latter would come at the cost of increased latency.
func Create(ctx context.Context, p CreateParams) (short string, err error) {
	if err := p.validate(); err != nil {
		return "", fmt.Errorf("invalid params: %v", err)
	}

	coll := p.DB.Database(p.Environment).Collection(Collection)

	// Generate a short URL and check for collisions.
	for length := defaultLength; length <= maxLength; length++ {
		short = randstr.New(length)

		filter := bson.M{"short": short}
		findContext, cancel := context.WithTimeout(ctx, findTimeout)
		defer cancel()

		// If the FindOne error is nil, it means a document
		// exists -- we have a collision.
		if err := coll.FindOne(findContext, filter).Err(); err == nil {
			log.Printf("collision: %s already exists", short)

			if length == maxLength {
				log.Printf("aborting: failed to generate unique short URL")
				return "", fmt.Errorf("failed to generate unique short url: %v", err)
			}
		}
	}

	// Insert a new ShortURL document.
	insertContext, cancel := context.WithTimeout(ctx, insertTimeout)
	defer cancel()

	if _, err := coll.InsertOne(insertContext, ShortURL{
		Created: time.Now(),
		Short:   short,
		URL:     p.LongURL,
	}); err != nil {
		return "", fmt.Errorf("failed to insert: %v", err)
	}

	return short, nil
}
