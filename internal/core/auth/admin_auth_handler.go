package auth

import (
	views "ecommerce/internal/admin/views"
	"ecommerce/internal/core/middleware"
	"ecommerce/internal/core/session"
	"ecommerce/internal/shared/httpx"

	"github.com/gofiber/fiber/v3"
)

type AdminAuthHandler struct {
	authService    *authService
	sessionManager *session.SessionManager
}

func NewAdminAuthHandler(authService *authService, sessionManager *session.SessionManager) *AdminAuthHandler {
	return &AdminAuthHandler{
		authService:    authService,
		sessionManager: sessionManager,
	}
}

// Mount, route'ları kaydeder. Admin'de kayıt yok: tek admin kullanıcısı
// açılışta EnsureAdmin ile env'den oluşturulur.
func (h *AdminAuthHandler) Mount(app *fiber.App) {
	withSess := middleware.WithSession(h.sessionManager, AdminScope)

	app.Get("/admin/login", h.loginPage, withSess)
	app.Post("/admin/login", h.login, withSess, middleware.CSRF)
	app.Post("/admin/logout", h.logout, withSess, middleware.CSRF)
}

func (h *AdminAuthHandler) loginPage(c fiber.Ctx) error {
	if middleware.SessionFromCtx(c) != nil {
		return httpx.Redirect(c, "/admin")
	}
	return httpx.Render(c, views.LoginPage(middleware.CSRFToken(c)))
}

func (h *AdminAuthHandler) login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Form(&req); err != nil {
		return httpx.Render(c, views.LoginForm(req.Email, "Geçersiz form verisi."))
	}

	// RequiredRole client'tan alınmaz (form:"-"), server-side set edilir
	req.RequiredRole = RoleAdmin
	if err := req.Validate(); err != nil {
		return httpx.Render(c, views.LoginForm(req.Email, err.Error()))
	}

	u, err := h.authService.Authenticate(c.Context(), req)
	if err != nil {
		return httpx.Render(c, views.LoginForm(req.Email, err.Error()))
	}

	h.sessionManager.Destroy(c.Context(), c, AdminScope)
	if _, err := h.sessionManager.Create(c.Context(), c, AdminScope, u.ID.Hex(), u.Email, u.Name, u.Role); err != nil {
		return httpx.Render(c, views.LoginForm(req.Email, "Giriş yapılamadı, lütfen tekrar deneyin."))
	}

	return httpx.Redirect(c, "/admin")
}

func (h *AdminAuthHandler) logout(c fiber.Ctx) error {
	h.sessionManager.Destroy(c.Context(), c, AdminScope)
	return httpx.Redirect(c, "/admin/login")
}
