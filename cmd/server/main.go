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
	"ecommerce/internal/core/session"
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

	// --- core servisler ---
	users := auth.NewAuthService(auth.NewAuthRepository(mongoDB))
	sessions := session.NewSessionManager(rdb, cfg.SessionIdleTimeout, cfg.SessionAbsoluteTimeout, cfg.IsProd())

	// ilk açılışta env'den gelen admin kullanıcısını oluştur
	if err := users.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		return err
	}

	// --- route'lar ---
	app := fiber.New(fiber.Config{AppName: cfg.StoreName})
	app.Use(logger.New())
	app.Use("/static", static.New("./static"))

	// her feature kendi handler'ıyla, sadece kendi bağımlılıklarını alarak mount edilir
	auth.NewAdminAuthHandler(users, sessions).Mount(app)
	adminhandlers.NewDashboard(sessions).Mount(app)
	storehandlers.NewHome(cfg.StoreName, sessions).Mount(app)
	auth.NewCustomerAuthHandler(users, sessions, cfg.StoreName).Mount(app)

	slog.Info("sunucu başlıyor",
		"addr", cfg.Addr,
		"storefront", "http://localhost"+cfg.Addr,
		"admin", "http://localhost"+cfg.Addr+"/admin",
	)
	return app.Listen(cfg.Addr)
}
