package auth

import (
	"github.com/gofiber/fiber/v3"

	"ecommerce/internal/admin/views"
	"ecommerce/internal/core/session"
	"ecommerce/internal/shared/httpx"
)

// AdminAuth — admin panel giriş/çıkış.
type AdminAuth struct {
	users *authService
}

func NewAdminAuth(users *authService) *AdminAuth {
	return &AdminAuth{users: users}
}

func (h *AdminAuth) Mount(app *fiber.App) {
	app.Get("/admin/login", h.loginPage)
	app.Post("/admin/login", h.login)
	app.Post("/admin/logout", h.logout)
}

func (h *AdminAuth) loginPage(c fiber.Ctx) error {
	if sess := session.FromCtx(c); sess != nil && sess.Role == RoleAdmin {
		return httpx.Redirect(c, "/admin")
	}
	return httpx.Render(c, views.LoginPage(session.Token(c)))
}

func (h *AdminAuth) login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Form(&req); err != nil {
		return httpx.Render(c, views.LoginForm("", ErrInvalidCredentials.Error()))
	}
	req.RequiredRole = RoleAdmin

	if err := req.Validate(); err != nil {
		return httpx.Render(c, views.LoginForm(req.Email, err.Error()))
	}

	u, err := h.users.Authenticate(c.Context(), req)
	if err != nil {
		return httpx.Render(c, views.LoginForm(req.Email, ErrInvalidCredentials.Error()))
	}

	if err := session.Login(c, &session.Session{
		UserID: u.ID.Hex(),
		Email:  u.Email,
		Name:   u.Name,
		Role:   u.Role,
	}); err != nil {
		return err
	}

	return httpx.Redirect(c, "/admin")
}

func (h *AdminAuth) logout(c fiber.Ctx) error {
	if err := session.Logout(c); err != nil {
		return err
	}
	return httpx.Redirect(c, "/admin/login")
}
