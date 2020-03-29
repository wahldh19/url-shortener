package visit

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateParams struct {
	DB          *mongo.Client
	Environment string
	ShortID     primitive.ObjectID
}

func (p CreateParams) validate() error {
	if p.DB == nil {
		return fmt.Errorf("missing db client")
	}
	if p.Environment == "" {
		return fmt.Errorf("invalid env")
	}
	if len(p.ShortID) == 0 {
		return fmt.Errorf("missing document id")
	}

	return nil
}

// Create creates a Visit document in MongoDB.
func Create(ctx context.Context, p CreateParams) error {
	if err := p.validate(); err != nil {
		return fmt.Errorf("invalid params: %v", err)
	}

	// Insert a new visit document.
	coll := p.DB.Database(p.Environment).Collection(Collection)
	insertContext, cancel := context.WithTimeout(ctx, insertTimeout)
	defer cancel()

	if _, err := coll.InsertOne(insertContext, Visit{
		ShortID: p.ShortID,
		Time:    time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to insert access record: %v", err)
	}

	return nil
}
