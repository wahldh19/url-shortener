package shorturl

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
)

type URLParams struct {
	DB          *mongo.Client
	Environment string
}

func (p URLParams) validate() error {
	if p.DB == nil {
		return fmt.Errorf("missing db client")
	}
	if p.Environment == "" {
		return fmt.Errorf("invalid env")
	}
	return nil
}

func GetURLCount(ctx context.Context, p URLParams) (URLCount int64, err error) {
	if err := p.validate(); err != nil {
		return 0, fmt.Errorf("invalid params: %v", err)
	}

	URLCount, err = p.DB.Database(p.Environment).Collection(Collection).CountDocuments(ctx, bson.D{}, nil)
	if err != nil {
		log.Print("Error: ", err)
		os.Exit(1)
	}
	return URLCount, nil
}
