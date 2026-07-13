// Package config, tüm uygulama ayarlarını environment variable'lardan okur.
// Yeni müşteride kod değiştirmeden .env dosyası ile her şey ayarlanabilir.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// HTTP
	Addr string // örn ":8080"
	Env  string // "dev" | "prod" — prod'da Secure cookie zorunlu

	// MongoDB
	MongoURI string
	MongoDB  string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Session
	SessionIdleTimeout     time.Duration // son istekten sonra ne kadar geçerli
	SessionAbsoluteTimeout time.Duration // login'den sonra maksimum ömür

	// İlk açılışta oluşturulacak admin kullanıcısı
	AdminEmail    string
	AdminPassword string

	// Payment — provider'a göre env'den okunur, kod değişmez
	PaymentProvider  string // "mock" | ileride: "iyzico", "stripe" ...
	PaymentAPIKey    string
	PaymentSecretKey string
	PaymentBaseURL   string

	// Storefront marka bilgisi (templ'lere config'ten gider)
	StoreName string
}

// Load, .env dosyasını (varsa) yükler ve env variable'lardan Config üretir.
func Load() (*Config, error) {
	loadDotEnv(".env") // yoksa sessizce geçer

	cfg := &Config{
		Addr: getEnv("ADDR", ":8080"),
		Env:  getEnv("APP_ENV", "dev"),

		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:  getEnv("MONGO_DB", "ecommerce"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		SessionIdleTimeout:     getEnvDuration("SESSION_IDLE_TIMEOUT", 30*time.Minute),
		SessionAbsoluteTimeout: getEnvDuration("SESSION_ABSOLUTE_TIMEOUT", 12*time.Hour),

		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@example.com"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),

		PaymentProvider:  getEnv("PAYMENT_PROVIDER", "mock"),
		PaymentAPIKey:    getEnv("PAYMENT_API_KEY", ""),
		PaymentSecretKey: getEnv("PAYMENT_SECRET_KEY", ""),
		PaymentBaseURL:   getEnv("PAYMENT_BASE_URL", ""),

		StoreName: getEnv("STORE_NAME", "Demo Store"),
	}

	if cfg.Env == "prod" && cfg.AdminPassword == "admin123" {
		return nil, fmt.Errorf("prod ortamında varsayılan ADMIN_PASSWORD kullanılamaz")
	}
	return cfg, nil
}

func (c *Config) IsProd() bool { return c.Env == "prod" }

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
