package shorturl

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CountParams struct {
	DB          *mongo.Client
	Environment string
}

func (p CountParams) validate() error {
	if p.DB == nil {
		return fmt.Errorf("missing db client")
	}
	if p.Environment == "" {
		return fmt.Errorf("invalid env")
	}
	return nil
}

func CountURLs(ctx context.Context, p CountParams) (count int64, err error) {
	if err := p.validate(); err != nil {
		return 0, fmt.Errorf("invalid params: %v", err)
	}

	count, err = p.DB.Database(
		p.Environment,
	).Collection(Collection).CountDocuments(ctx, bson.D{}, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %v", err)
	}

	return count, nil
}
