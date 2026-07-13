package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewMongo, MongoDB'ye bağlanır ve veritabanı handle'ı döner.
func NewMongo(ctx context.Context, uri, dbName string) (*mongo.Database, func(), error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, nil, err
	}
	closeFn := func() {
		_ = client.Disconnect(context.Background())
	}
	return client.Database(dbName), closeFn, nil
}
