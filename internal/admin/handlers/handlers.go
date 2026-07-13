// Package handlers (admin) — SABİT, müşteriye göre değişmez.
package handlers

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v3"

	views "ecommerce/internal/admin/views"
	"ecommerce/internal/core/auth"
	"ecommerce/internal/core/order"
	"ecommerce/internal/core/product"
	"ecommerce/internal/shared/httpx"
	"ecommerce/internal/shared/middleware"
)

type Handlers struct {
	products *product.Service
	orders   *order.Service
	users    *auth.Service
	sessions *auth.SessionManager
}

func New(products *product.Service, orders *order.Service, users *auth.Service, sessions *auth.SessionManager) *Handlers {
	return &Handlers{products: products, orders: orders, users: users, sessions: sessions}
}

// Mount, admin route'larını uygulamaya bağlar.
// Login dışındaki her şey RequireRole(admin) + CSRF ile korunur.
func (h *Handlers) Mount(app *fiber.App) {
	app.Get("/admin/login", h.loginPage)
	app.Post("/admin/login", h.login, middleware.CSRF) // login öncesi CSRF: double-submit cookie

	admin := app.Group("/admin",
		middleware.RequireRole(h.sessions, auth.AdminScope, auth.RoleAdmin),
		middleware.CSRF,
	)
	admin.Post("/logout", h.logout)
	admin.Get("/", h.dashboard)

	admin.Get("/products", h.productList)
	admin.Get("/products/new", h.productNew)
	admin.Post("/products", h.productCreate)
	admin.Get("/products/:id/edit", h.productEdit)
	admin.Put("/products/:id", h.productUpdate)
	admin.Get("/products/:id/row", h.productRow)
	admin.Get("/products/:id/edit-row", h.productEditRow)
	admin.Put("/products/:id/inline", h.productInlineUpdate)
	admin.Delete("/products/:id", h.productDelete)

	admin.Get("/orders", h.orderList)
	admin.Post("/orders/:id/status", h.orderStatus)

	admin.Get("/users", h.userList)
}

// --- auth ---

func (h *Handlers) loginPage(c fiber.Ctx) error {
	return httpx.Render(c, views.LoginPage(middleware.CSRFToken(c), ""))
}

func (h *Handlers) login(c fiber.Ctx) error {
	u, err := h.users.Authenticate(c.Context(), c.FormValue("email"), c.FormValue("password"), auth.RoleAdmin)
	if err != nil {
		return httpx.Render(c, views.LoginPage(middleware.CSRFToken(c), "E-posta veya şifre hatalı"))
	}
	// session fixation önlemi: varsa eski session'ı yok et, yeni ID üret
	h.sessions.Destroy(c.Context(), c, auth.AdminScope)
	if _, err := h.sessions.Create(c.Context(), c, auth.AdminScope, u); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "oturum oluşturulamadı")
	}
	return c.Redirect().Status(fiber.StatusSeeOther).To("/admin")
}

func (h *Handlers) logout(c fiber.Ctx) error {
	h.sessions.Destroy(c.Context(), c, auth.AdminScope)
	return httpx.Redirect(c, "/admin/login")
}

// --- dashboard ---

func (h *Handlers) dashboard(c fiber.Ctx) error {
	stats, err := h.orders.Stats(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	count, _ := h.products.Count(c.Context())
	recent, _ := h.orders.List(c.Context())
	if len(recent) > 10 {
		recent = recent[:10]
	}
	s := middleware.SessionFromCtx(c)
	return httpx.Render(c, views.Dashboard(s.CSRFToken, s.Name, stats, count, recent))
}

// --- products ---

func (h *Handlers) productList(c fiber.Ctx) error {
	list, err := h.products.List(c.Context(), product.ListFilter{})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	s := middleware.SessionFromCtx(c)
	return httpx.Render(c, views.ProductsPage(s.CSRFToken, s.Name, list))
}

func (h *Handlers) productNew(c fiber.Ctx) error {
	s := middleware.SessionFromCtx(c)
	return httpx.Render(c, views.ProductFormPage(s.CSRFToken, s.Name, nil, ""))
}

func (h *Handlers) productEdit(c fiber.Ctx) error {
	p, err := h.products.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound)
	}
	s := middleware.SessionFromCtx(c)
	return httpx.Render(c, views.ProductFormPage(s.CSRFToken, s.Name, p, ""))
}

func parseProductForm(c fiber.Ctx) product.CreateInput {
	price, err := strconv.ParseFloat(c.FormValue("price"), 64)
	if err != nil {
		price = 0
	}
	stock, _ := strconv.Atoi(c.FormValue("stock"))
	return product.CreateInput{
		Name:        c.FormValue("name"),
		Description: c.FormValue("description"),
		PriceCents:  int64(math.Round(price * 100)),
		ImageURL:    c.FormValue("image_url"),
		Stock:       stock,
		Active:      c.FormValue("active") == "on",
	}
}

func (h *Handlers) productCreate(c fiber.Ctx) error {
	in := parseProductForm(c)
	if _, err := h.products.Create(c.Context(), in); err != nil {
		return httpx.Render(c, views.ProductForm(nil, err.Error()))
	}
	return httpx.Redirect(c, "/admin/products")
}

func (h *Handlers) productUpdate(c fiber.Ctx) error {
	in := parseProductForm(c)
	p, err := h.products.Update(c.Context(), c.Params("id"), in)
	if err != nil {
		return httpx.Render(c, views.ProductForm(p, err.Error()))
	}
	return httpx.Redirect(c, "/admin/products")
}

func (h *Handlers) productRow(c fiber.Ctx) error {
	p, err := h.products.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound)
	}
	return httpx.Render(c, views.ProductRow(*p))
}

func (h *Handlers) productEditRow(c fiber.Ctx) error {
	p, err := h.products.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound)
	}
	return httpx.Render(c, views.ProductRowEdit(*p))
}

// productInlineUpdate — tablo içi hızlı düzenleme: sadece ad/fiyat/stok/aktiflik.
func (h *Handlers) productInlineUpdate(c fiber.Ctx) error {
	p, err := h.products.GetByID(c.Context(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound)
	}
	price, err := strconv.ParseFloat(c.FormValue("price"), 64)
	if err != nil {
		price = float64(p.PriceCents) / 100
	}
	stock, err := strconv.Atoi(c.FormValue("stock"))
	if err != nil {
		stock = p.Stock
	}
	in := product.CreateInput{
		Name:        c.FormValue("name"),
		Description: p.Description,
		PriceCents:  int64(math.Round(price * 100)),
		ImageURL:    p.ImageURL,
		Stock:       stock,
		Active:      c.FormValue("active") == "on",
	}
	updated, err := h.products.Update(c.Context(), c.Params("id"), in)
	if err != nil {
		return httpx.Render(c, views.ProductRowEdit(*p))
	}
	return httpx.Render(c, views.ProductRow(*updated))
}

func (h *Handlers) productDelete(c fiber.Ctx) error {
	if err := h.products.Delete(c.Context(), c.Params("id")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendString("") // boş cevap: htmx satırı DOM'dan siler
}

// --- orders ---

func (h *Handlers) orderList(c fiber.Ctx) error {
	list, err := h.orders.List(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	s := middleware.SessionFromCtx(c)
	return httpx.Render(c, views.OrdersPage(s.CSRFToken, s.Name, list))
}

func (h *Handlers) orderStatus(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.orders.UpdateStatus(c.Context(), id, c.FormValue("status")); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	o, err := h.orders.GetByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound)
	}
	return httpx.Render(c, views.OrderRow(*o))
}

// --- users ---

func (h *Handlers) userList(c fiber.Ctx) error {
	list, err := h.users.ListUsers(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	s := middleware.SessionFromCtx(c)
	return httpx.Render(c, views.UsersPage(s.CSRFToken, s.Name, list))
}
