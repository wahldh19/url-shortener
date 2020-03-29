package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	// connectTimeout is how long to wait before giving up on
	// connecting to the MongoDB cluster.
	connectTimeout = 5 * time.Second

	// pingTimeout is how long to wait before giving up on verifying
	// that a connection has been established to a MongoDB server.
	pingTimeout = 3 * time.Second
)

func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	log.Println("connecting to db")

	clientOptions := options.Client().ApplyURI(uri)
	connectCtx, connectCancel := context.WithTimeout(ctx, connectTimeout)
	defer connectCancel()

	client, err := mongo.Connect(connectCtx, clientOptions)
	if err != nil {
		return nil, err
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, pingTimeout)
	defer pingCancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, err
	}

	log.Println("connected to db")

	return client, nil
}
