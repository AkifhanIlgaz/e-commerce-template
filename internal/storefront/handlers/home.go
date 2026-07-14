package handlers

import (
	"github.com/gofiber/fiber/v3"

	"ecommerce/internal/core/session"
	"ecommerce/internal/shared/httpx"
	"ecommerce/internal/storefront/views"
)

// Home — ana sayfa. Ürün listesi eklendiğinde product servisi
// bağımlılık olarak buraya gelecek.
type Home struct {
	storeName string
}

func NewHome(storeName string) *Home {
	return &Home{storeName: storeName}
}

func (h *Home) Mount(app *fiber.App) {
	app.Get("/", h.home)
}

func (h *Home) home(c fiber.Ctx) error {
	return httpx.Render(c, views.Home(h.storeName, session.Token(c), session.FromCtx(c)))
}
