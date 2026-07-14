package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"

	"ecommerce/internal/admin/handlers"
	"ecommerce/internal/config"
	"ecommerce/internal/core/auth"
	"ecommerce/internal/core/session"
	"ecommerce/internal/shared/db"
	storefront "ecommerce/internal/storefront/handlers"
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

	sessionStorage, closeRedis, err := db.NewRedisStorage(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		return err
	}
	defer closeRedis()

	// --- core servisler ---
	users := auth.NewAuthService(auth.NewAuthRepository(mongoDB))

	// ilk açılışta env'den gelen admin kullanıcısını oluştur
	if err := users.EnsureAdmin(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		return err
	}

	// --- route'lar ---
	app := fiber.New(fiber.Config{AppName: cfg.StoreName})
	app.Use(logger.New())
	app.Use("/static", static.New("./static"))

	// session + csrf tüm route'larda global; auth ayrımı rol middleware'iyle
	sessionMW, sessionStore := session.New(sessionStorage, cfg.SessionIdleTimeout, cfg.SessionAbsoluteTimeout)
	app.Use(sessionMW)
	app.Use(session.CSRF(sessionStore, cfg.SessionIdleTimeout))

	// her feature kendi handler'ıyla, sadece kendi bağımlılıklarını alarak mount edilir
	storefront.NewHome(cfg.StoreName).Mount(app)
	auth.NewCustomerAuth(cfg.StoreName, users).Mount(app)
	auth.NewAdminAuth(users).Mount(app)
	handlers.NewDashboard().Mount(app)

	slog.Info("sunucu başlıyor",
		"addr", cfg.Addr,
		"storefront", "http://localhost"+cfg.Addr,
		"admin", "http://localhost"+cfg.Addr+"/admin",
	)

	return app.Listen(cfg.Addr)
}
