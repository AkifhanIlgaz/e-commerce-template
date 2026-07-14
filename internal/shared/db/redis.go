package db

import (
	"fmt"

	fiberredis "github.com/gofiber/storage/redis/v3"
)

// NewRedisStorage, fiber session middleware'inin kullanacağı Redis storage'ı
// kurar. Paket bağlantı hatasında panic'lediği için error'a çevrilir.
func NewRedisStorage(addr, password string, dbNum int) (storage *fiberredis.Storage, closeFn func(), err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("redis bağlantısı kurulamadı: %v", r)
		}
	}()

	storage = fiberredis.New(fiberredis.Config{
		Addrs:    []string{addr},
		Password: password,
		Database: dbNum,
	})

	return storage, func() { _ = storage.Close() }, nil
}
