// Package session, fiber'ın built-in session ve csrf middleware'lerini
// proje ayarlarıyla (Redis storage, __Host- cookie'ler) kurar ve session
// verisini view'ların kullandığı tipli Session struct'ına çevirir.
package session

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	fibersession "github.com/gofiber/fiber/v3/middleware/session"
)

// Session verisinde kullanılan key'ler.
const (
	keyUserSession = "user_session"
)

// Session, giriş yapmış kullanıcının view'lara giden özeti.
// Kullanıcı giriş yapmamışsa FromCtx nil döner.
type Session struct {
	UserID string `redis:"user_id"`
	Email  string `redis:"email"`
	Name   string `redis:"name"`
	Role   string `redis:"role"`
}

// New, Redis storage üzerinde çalışan session middleware'ini ve store'u kurar.
// __Host- prefix'i Secure + Path=/ + Domain'siz cookie zorunlu kılar.
func New(storage fiber.Storage, idleTimeout, absoluteTimeout time.Duration) (fiber.Handler, *fibersession.Store) {
	mw, store := fibersession.NewWithStore(fibersession.Config{
		Storage:         storage,
		IdleTimeout:     idleTimeout,
		AbsoluteTimeout: absoluteTimeout,
		Extractor:       extractors.FromCookie("__Host-session"),
		CookiePath:      "/",
		CookieSecure:    true,
		CookieHTTPOnly:  true,
		CookieSameSite:  "Lax",
	})
	store.RegisterType(&Session{})

	return mw, store
}

// CSRF, session store'a bağlı csrf middleware'ini kurar. Token, layout'taki
// meta tag üzerinden htmx isteklerine X-CSRF-Token header'ı olarak eklenir;
// default extractor da header'dan okur.
func CSRF(store *fibersession.Store, idleTimeout time.Duration) fiber.Handler {
	return csrf.New(csrf.Config{
		Session:        store,
		CookieName:     "__Host-csrf_",
		CookiePath:     "/",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
		IdleTimeout:    idleTimeout,
	})
}

// Token, view'lara gömülecek csrf token'ını döner.
func Token(c fiber.Ctx) string {
	return csrf.TokenFromContext(c)
}

// FromCtx, session'daki kullanıcıyı okur; giriş yoksa nil döner.
func FromCtx(c fiber.Ctx) *Session {
	m := fibersession.FromContext(c)
	if m == nil {
		return nil
	}

	session, ok := m.Get(keyUserSession).(*Session)
	if !ok {
		return nil
	}

	return session
}

// Login, session fixation'a karşı session ID'yi yeniler ve kullanıcıyı yazar.
func Login(c fiber.Ctx, session *Session) error {
	m := fibersession.FromContext(c)

	if err := m.Regenerate(); err != nil {
		return err
	}

	m.Set(keyUserSession, session)

	return nil
}

// Logout, session'ı ve içindeki csrf token'ını tamamen yok eder.
func Logout(c fiber.Ctx) error {
	return fibersession.FromContext(c).Reset()
}
