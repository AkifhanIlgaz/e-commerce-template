package auth

import (
	"ecommerce/internal/core/middleware"
	"ecommerce/internal/core/session"
	"ecommerce/internal/shared/httpx"
	"ecommerce/internal/storefront/views"

	"github.com/gofiber/fiber/v3"
)

type CustomerAuthHandler struct {
	storeName      string
	authService    *authService
	sessionManager *session.SessionManager
}

func NewCustomerAuthHandler(authService *authService, sessionManager *session.SessionManager, storeName string) *CustomerAuthHandler {
	return &CustomerAuthHandler{
		storeName:      storeName,
		authService:    authService,
		sessionManager: sessionManager,
	}
}

// Mount, route'ları kaydeder. Fiber v3'te middleware'ler handler'dan SONRA
// verilir ama handler'dan ÖNCE çalışır. Group("/") kullanmıyoruz çünkü
// prefix tabanlı middleware /admin route'larına da bulaşırdı.
func (h *CustomerAuthHandler) Mount(app *fiber.App) {
	withSess := middleware.WithSession(h.sessionManager, StoreScope)
	requireCustomer := middleware.RequireRole(h.sessionManager, StoreScope, RoleCustomer)

	app.Get("/login", h.loginPage, withSess)
	app.Post("/login", h.login, withSess, middleware.CSRF)
	app.Get("/register", h.registerPage, withSess)
	app.Post("/register", h.register, withSess, middleware.CSRF)
	app.Post("/logout", h.logout, withSess, middleware.CSRF)

	app.Get("/account", h.account, requireCustomer)
}

func (h *CustomerAuthHandler) loginPage(c fiber.Ctx) error {
	if middleware.SessionFromCtx(c) != nil {
		return httpx.Redirect(c, "/account")
	}
	return httpx.Render(c, views.LoginPage(h.storeName, middleware.CSRFToken(c)))
}

func (h *CustomerAuthHandler) login(c fiber.Ctx) error {
	email := c.FormValue("email")
	u, err := h.authService.Authenticate(c.Context(), LoginRequest{
		Email:        email,
		Password:     c.FormValue("password"),
		RequiredRole: RoleCustomer,
	})
	if err != nil {
		return httpx.Render(c, views.LoginForm(email, err.Error()))
	}
	// session fixation önlemi: varsa eski session'ı sil, yeni ID üret
	h.sessionManager.Destroy(c.Context(), c, StoreScope)
	if _, err := h.sessionManager.Create(c.Context(), c, StoreScope, u.ID.Hex(), u.Email, u.Name, u.Role); err != nil {
		return httpx.Render(c, views.LoginForm(email, "Giriş yapılamadı, lütfen tekrar deneyin."))
	}
	return httpx.Redirect(c, "/account")
}

func (h *CustomerAuthHandler) registerPage(c fiber.Ctx) error {
	if middleware.SessionFromCtx(c) != nil {
		return httpx.Redirect(c, "/account")
	}
	return httpx.Render(c, views.RegisterPage(h.storeName, middleware.CSRFToken(c)))
}

func (h *CustomerAuthHandler) register(c fiber.Ctx) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	u, err := h.authService.Register(c.Context(), RegisterRequest{
		Email:    email,
		Password: c.FormValue("password"),
		Name:     name,
		Role:     RoleCustomer,
	})
	if err != nil {
		return httpx.Render(c, views.RegisterForm(name, email, err.Error()))
	}
	// kayıttan sonra otomatik giriş
	if _, err := h.sessionManager.Create(c.Context(), c, StoreScope, u.ID.Hex(), u.Email, u.Name, u.Role); err != nil {
		return httpx.Redirect(c, "/login")
	}
	return httpx.Redirect(c, "/account")
}

func (h *CustomerAuthHandler) logout(c fiber.Ctx) error {
	h.sessionManager.Destroy(c.Context(), c, StoreScope)
	return httpx.Redirect(c, "/")
}

func (h *CustomerAuthHandler) account(c fiber.Ctx) error {
	sess := middleware.SessionFromCtx(c)
	return httpx.Render(c, views.AccountPage(h.storeName, middleware.CSRFToken(c), sess))
}
