package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"

	adminhandlers "ecommerce/internal/admin/handlers"
	"ecommerce/internal/config"
	"ecommerce/internal/core/auth"
	"ecommerce/internal/core/cart"
	"ecommerce/internal/core/checkout"
	"ecommerce/internal/core/order"
	"ecommerce/internal/core/payment"
	"ecommerce/internal/core/product"
	"ecommerce/internal/shared/db"
	storehandlers "ecommerce/internal/storefront/handlers"
)

func main() {
	if err := run(); err != nil {
		slog.Error("başlatılamadı", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// --- altyapı bağlantıları ---
	mongoDB, closeMongo, err := db.NewMongo(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		return err
	}
	defer closeMongo()

	rdb, closeRedis, err := db.NewRedis(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		return err
	}
	defer closeRedis()

	// --- core servisler (admin + storefront tarafından PAYLAŞILIR) ---
	products := product.NewService(product.NewRepository(mongoDB))
	orders := order.NewService(order.NewRepository(mongoDB))
	users := auth.NewService(auth.NewUserRepository(mongoDB))
	carts := cart.NewService(rdb, products, cfg.IsProd())
	sessions := auth.NewSessionManager(rdb, cfg.SessionIdleTimeout, cfg.SessionAbsoluteTimeout, cfg.IsProd())

	provider, err := payment.NewProvider(cfg)
	if err != nil {
		return err
	}
	co := checkout.NewService(carts, orders, products, provider)

	// --- ilk açılış: admin kullanıcısı + örnek ürünler ---
	if err := users.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		return err
	}
	if err := seedProducts(ctx, products); err != nil {
		return err
	}

	// --- route'lar ---
	app := fiber.New(fiber.Config{AppName: cfg.StoreName})
	app.Use(logger.New())
	app.Use("/static", static.New("./static"))

	adminhandlers.New(products, orders, users, sessions).Mount(app)
	storehandlers.New(cfg.StoreName, products, carts, orders, co, users, sessions, provider).Mount(app)

	slog.Info("sunucu başlıyor",
		"addr", cfg.Addr,
		"storefront", "http://localhost"+cfg.Addr,
		"admin", "http://localhost"+cfg.Addr+"/admin",
		"payment_provider", provider.Name(),
	)
	return app.Listen(cfg.Addr)
}

// seedProducts, veritabanı boşsa demo ürünler ekler — yeni klonda site boş görünmesin diye.
func seedProducts(ctx context.Context, products *product.Service) error {
	n, err := products.Count(ctx)
	if err != nil || n > 0 {
		return err
	}
	demo := []product.CreateInput{
		{Name: "Klasik Beyaz Tişört", Description: "%100 pamuk, günlük kullanım için rahat kesim.", PriceCents: 29990, Stock: 50, Active: true, ImageURL: "https://picsum.photos/seed/tshirt/600"},
		{Name: "Deri Cüzdan", Description: "El yapımı hakiki deri cüzdan, 8 kart bölmesi.", PriceCents: 79990, Stock: 20, Active: true, ImageURL: "https://picsum.photos/seed/wallet/600"},
		{Name: "Seramik Kupa", Description: "350ml, bulaşık makinesinde yıkanabilir.", PriceCents: 19990, Stock: 100, Active: true, ImageURL: "https://picsum.photos/seed/mug/600"},
		{Name: "Spor Ayakkabı", Description: "Hafif taban, nefes alan kumaş.", PriceCents: 149990, Stock: 30, Active: true, ImageURL: "https://picsum.photos/seed/shoe/600"},
	}
	for _, in := range demo {
		if _, err := products.Create(ctx, in); err != nil {
			return err
		}
	}
	slog.Info("demo ürünler eklendi", "adet", len(demo))
	return nil
}
