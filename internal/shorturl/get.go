package shorturl

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type GetParams struct {
	DB          *mongo.Client
	Environment string
	Short       string
}

func (p GetParams) validate() error {
	if p.DB == nil {
		return fmt.Errorf("missing db client")
	}
	if p.Environment == "" {
		return fmt.Errorf("invalid env")
	}
	if p.Short == "" {
		return fmt.Errorf("missing short")
	}

	return nil
}

// Get retrieves a short URL document from the database.
func Get(ctx context.Context, p GetParams) (short *ShortURL, err error) {
	if err := p.validate(); err != nil {
		return nil, fmt.Errorf("invalid params: %v", err)
	}

	coll := p.DB.Database(p.Environment).Collection(Collection)

	findContext, cancel := context.WithTimeout(ctx, findTimeout)
	defer cancel()

	filter := bson.M{"short": p.Short}
	if err := coll.FindOne(findContext, filter).Decode(&short); err != nil {
		return nil, fmt.Errorf("failed to find: %v", err)
	}

	return short, nil
}
