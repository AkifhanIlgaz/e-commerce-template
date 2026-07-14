package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// NewMongo, MongoDB'ye bağlanır ve veritabanı handle'ı döner.
func NewMongo(ctx context.Context, uri, dbName string) (*mongo.Database, func(), error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, nil, err
	}
	closeFn := func() {
		_ = client.Disconnect(context.Background())
	}
	return client.Database(dbName), closeFn, nil
}
