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
	email := c.FormValue("email")
	u, err := h.authService.Authenticate(c.Context(), LoginRequest{
		Email:        email,
		Password:     c.FormValue("password"),
		RequiredRole: RoleAdmin,
	})
	if err != nil {
		return httpx.Render(c, views.LoginForm(email, err.Error()))
	}
	// session fixation önlemi: varsa eski session'ı sil, yeni ID üret
	h.sessionManager.Destroy(c.Context(), c, AdminScope)
	if _, err := h.sessionManager.Create(c.Context(), c, AdminScope, u.ID.Hex(), u.Email, u.Name, u.Role); err != nil {
		return httpx.Render(c, views.LoginForm(email, "Giriş yapılamadı, lütfen tekrar deneyin."))
	}
	return httpx.Redirect(c, "/admin")
}

func (h *AdminAuthHandler) logout(c fiber.Ctx) error {
	h.sessionManager.Destroy(c.Context(), c, AdminScope)
	return httpx.Redirect(c, "/admin/login")
}
