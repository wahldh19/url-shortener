package shorturl

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UrlParams struct {
	DB          *mongo.Client
	Environment string
	Query       string
}

func (p UrlParams) validate() error {
	if p.DB == nil {
		return fmt.Errorf("missing db client")
	}
	if p.Environment == "" {
		return fmt.Errorf("invalid env")
	}
	if p.Query == "" {
		return fmt.Errorf("missing Query")
	}

	return nil
}

func GetUrlCount(ctx context.Context, p UrlParams) (UrlCount int64, err error) {
	if err := p.validate(); err != nil {
		return 0, fmt.Errorf("invalid params: %v", err)
	}
	UrlCount, err = p.DB.Database(p.Environment).Collection(Collection).CountDocuments(ctx, bson.D{}, nil)
	return UrlCount, nil
}
