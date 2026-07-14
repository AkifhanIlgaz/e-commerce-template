package handlers

import (
	"ecommerce/internal/admin/views"
	"ecommerce/internal/core/auth"
	"ecommerce/internal/core/middleware"
	"ecommerce/internal/core/session"
	"ecommerce/internal/shared/httpx"

	"github.com/gofiber/fiber/v3"
)

// Dashboard — panel ana sayfası. İstatistikler geldiğinde order/product
// servisleri bağımlılık olarak buraya gelecek.
type Dashboard struct {
	sessions *session.SessionManager
}

func NewDashboard(sessions *session.SessionManager) *Dashboard {
	return &Dashboard{sessions: sessions}
}

func (h *Dashboard) Mount(app *fiber.App) {
	app.Get("/admin", h.dashboard, middleware.RequireRole(h.sessions, session.AdminScope, auth.RoleAdmin))
}

func (h *Dashboard) dashboard(c fiber.Ctx) error {
	sess := middleware.SessionFromCtx(c)
	return httpx.Render(c, views.Dashboard(middleware.CSRFToken(c), sess))
}
