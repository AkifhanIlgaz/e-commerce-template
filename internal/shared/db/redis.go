package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedis, Redis'e bağlanır ve client döner.
func NewRedis(ctx context.Context, addr, password string, dbNum int) (*redis.Client, func(), error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbNum,
	})
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		return nil, nil, err
	}
	closeFn := func() { _ = client.Close() }
	return client, closeFn, nil
}
