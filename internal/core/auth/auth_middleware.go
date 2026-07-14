package auth

import (
	"github.com/gofiber/fiber/v3"

	"ecommerce/internal/core/session"
	"ecommerce/internal/shared/httpx"
)

// RequireAuth, giriş yapmamış kullanıcıyı müşteri login sayfasına yollar.
func RequireAuth(c fiber.Ctx) error {
	if session.FromCtx(c) == nil {
		return httpx.Redirect(c, "/login")
	}

	return c.Next()
}

// RequireAdmin, admin rolü olmayan herkesi admin login sayfasına yollar.
// Tek session store kullanıldığı için ayrım rol kontrolüyle yapılır.
func RequireAdmin(c fiber.Ctx) error {
	sess := session.FromCtx(c)
	if sess == nil || sess.Role != RoleAdmin {
		return httpx.Redirect(c, "/admin/login")
	}
	return c.Next()
}
