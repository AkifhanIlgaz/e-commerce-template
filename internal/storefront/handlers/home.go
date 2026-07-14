package handlers

import (
	"github.com/gofiber/fiber/v3"

	"ecommerce/internal/core/middleware"
	"ecommerce/internal/core/session"
	"ecommerce/internal/shared/httpx"
	"ecommerce/internal/storefront/views"
)

// Home — ana sayfa. Ürün listesi eklendiğinde product servisi
// bağımlılık olarak buraya gelecek.
type Home struct {
	storeName string
	sessions  *session.SessionManager
}

func NewHome(storeName string, sessions *session.SessionManager) *Home {
	return &Home{storeName: storeName, sessions: sessions}
}

func (h *Home) Mount(app *fiber.App) {
	app.Get("/", h.home, middleware.WithSession(h.sessions, session.StoreScope))
}

func (h *Home) home(c fiber.Ctx) error {
	sess := middleware.SessionFromCtx(c)
	return httpx.Render(c, views.Home(h.storeName, middleware.CSRFToken(c), sess))
}
