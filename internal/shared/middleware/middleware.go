package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"

	"github.com/gofiber/fiber/v3"

	"ecommerce/internal/core/auth"
)

const sessionLocalKey = "session"

// SessionFromCtx, WithSession/RequireRole middleware'inin koyduğu session'ı döner (yoksa nil).
func SessionFromCtx(c fiber.Ctx) *auth.Session {
	s, _ := c.Locals(sessionLocalKey).(*auth.Session)
	return s
}

// WithSession, geçerli bir session varsa Locals'a koyar; yoksa isteği yine de geçirir.
// Storefront'ta kullanılır: misafir de gezebilir, giriş yaptıysa adı görünür.
func WithSession(sm *auth.SessionManager, scope auth.Scope) fiber.Handler {
	return func(c fiber.Ctx) error {
		if sess, err := sm.Get(c.Context(), c, scope); err == nil {
			c.Locals(sessionLocalKey, sess)
		}
		return c.Next()
	}
}

// RequireRole, session yoksa veya rol eşleşmiyorsa login sayfasına yönlendirir.
// Admin paneli RequireRole(sm, auth.AdminScope, auth.RoleAdmin) ile korunur.
func RequireRole(sm *auth.SessionManager, scope auth.Scope, role string) fiber.Handler {
	return func(c fiber.Ctx) error {
		sess, err := sm.Get(c.Context(), c, scope)
		if err != nil || sess.Role != role {
			if c.Get("HX-Request") == "true" {
				c.Set("HX-Redirect", scope.LoginURL)
				return c.Status(fiber.StatusUnauthorized).SendString("")
			}
			return c.Redirect().Status(fiber.StatusSeeOther).To(scope.LoginURL)
		}
		c.Locals(sessionLocalKey, sess)
		return c.Next()
	}
}

// --- CSRF ---
//
// Giriş yapmış kullanıcıda token session'da tutulur. Misafirde (storefront
// sepeti) double-submit cookie kullanılır: HttpOnly cookie'deki token,
// sayfadaki <meta name="csrf-token"> üzerinden htmx'in X-CSRF-Token
// header'ına yazılır ve ikisi karşılaştırılır.

const csrfCookieName = "csrf_token"

// CSRFToken, sayfa render edilirken meta tag'e konacak token'ı döner.
// Misafirde cookie yoksa oluşturur.
func CSRFToken(c fiber.Ctx) string {
	if sess := SessionFromCtx(c); sess != nil {
		return sess.CSRFToken
	}
	if v := c.Cookies(csrfCookieName); v != "" {
		return v
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	c.Cookie(&fiber.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
	return token
}

// CSRF, güvenli olmayan metodlarda (POST/PUT/PATCH/DELETE) token doğrular.
// Session middleware'inden SONRA zincirlenmeli.
func CSRF(c fiber.Ctx) error {
	switch c.Method() {
	case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
		return c.Next()
	}
	got := c.Get("X-CSRF-Token")
	if got == "" {
		got = c.FormValue("csrf_token")
	}
	var want string
	if sess := SessionFromCtx(c); sess != nil {
		want = sess.CSRFToken
	} else {
		want = c.Cookies(csrfCookieName)
	}
	if want == "" || subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
		return c.Status(fiber.StatusForbidden).SendString("CSRF token geçersiz")
	}
	return c.Next()
}
