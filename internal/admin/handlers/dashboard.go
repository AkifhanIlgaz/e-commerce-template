package handlers

import (
	"ecommerce/internal/admin/views"
	"ecommerce/internal/core/auth"
	"ecommerce/internal/core/session"
	"ecommerce/internal/shared/httpx"

	"github.com/gofiber/fiber/v3"
)

// Dashboard — panel ana sayfası. İstatistikler geldiğinde order/product
// servisleri bağımlılık olarak buraya gelecek.
type Dashboard struct{}

func NewDashboard() *Dashboard {
	return &Dashboard{}
}

func (h *Dashboard) Mount(app *fiber.App) {
	app.Get("/admin", auth.RequireAdmin, h.dashboard)
}

func (h *Dashboard) dashboard(c fiber.Ctx) error {
	return httpx.Render(c, views.Dashboard(session.Token(c), session.FromCtx(c)))
}
