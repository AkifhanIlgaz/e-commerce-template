package auth

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"ecommerce/internal/core/session"
	"ecommerce/internal/shared/httpx"
	"ecommerce/internal/storefront/views"
)

// CustomerAuth — storefront giriş/kayıt/çıkış ve hesap sayfası.
type CustomerAuth struct {
	storeName string
	users     *authService
}

func NewCustomerAuth(storeName string, users *authService) *CustomerAuth {
	return &CustomerAuth{storeName: storeName, users: users}
}

func (h *CustomerAuth) Mount(app *fiber.App) {
	app.Get("/login", h.loginPage)
	app.Post("/login", h.login)
	app.Get("/register", h.registerPage)
	app.Post("/register", h.register)
	app.Post("/logout", h.logout)
	app.Get("/account", RequireAuth, h.account)
}

func (h *CustomerAuth) loginPage(c fiber.Ctx) error {
	if session.FromCtx(c) != nil {
		return httpx.Redirect(c, "/")
	}
	return httpx.Render(c, views.LoginPage(h.storeName, session.Token(c)))
}

func (h *CustomerAuth) login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Form(&req); err != nil {
		return httpx.Render(c, views.LoginForm("", ErrInvalidCredentials.Error()))
	}
	req.RequiredRole = RoleCustomer

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

	return httpx.Redirect(c, "/")
}

func (h *CustomerAuth) registerPage(c fiber.Ctx) error {
	if session.FromCtx(c) != nil {
		return httpx.Redirect(c, "/")
	}
	return httpx.Render(c, views.RegisterPage(h.storeName, session.Token(c)))
}

func (h *CustomerAuth) register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Form(&req); err != nil {
		return httpx.Render(c, views.RegisterForm("", "", "Form bilgileri geçersiz"))
	}
	req.Role = RoleCustomer

	if err := req.Validate(); err != nil {
		return httpx.Render(c, views.RegisterForm(req.Name, req.Email, err.Error()))
	}

	u, err := h.users.Register(c.Context(), req)
	if err != nil {
		msg := "Kayıt oluşturulamadı"
		if errors.Is(err, ErrEmailTaken) {
			msg = err.Error()
		}
		return httpx.Render(c, views.RegisterForm(req.Name, req.Email, msg))
	}

	if err := session.Login(c, &session.Session{
		UserID: u.ID.Hex(),
		Email:  u.Email,
		Name:   u.Name,
		Role:   u.Role,
	}); err != nil {
		return err
	}

	return httpx.Redirect(c, "/")
}

func (h *CustomerAuth) logout(c fiber.Ctx) error {
	if err := session.Logout(c); err != nil {
		return err
	}

	return httpx.Redirect(c, "/")
}

func (h *CustomerAuth) account(c fiber.Ctx) error {
	return httpx.Render(c, views.AccountPage(h.storeName, session.Token(c), session.FromCtx(c)))
}
