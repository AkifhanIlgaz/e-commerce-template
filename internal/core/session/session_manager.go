package session

import (
	"context"
	"crypto/rand"

	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

// SessionManager, Redis üzerinde scope'lu session yönetimi yapar.
type SessionManager struct {
	rdb         *redis.Client
	idleTTL     time.Duration
	absoluteTTL time.Duration
	secure      bool // prod'da true
}

func NewSessionManager(rdb *redis.Client, idle, absolute time.Duration, secure bool) *SessionManager {
	return &SessionManager{rdb: rdb, idleTTL: idle, absoluteTTL: absolute, secure: secure}
}

// Create, yeni session oluşturur ve cookie'yi yazar.
// Session fixation önlemi: login'de HER ZAMAN yeni ID üretilir,
// varsa eski session silinir (login handler'ları önce Destroy çağırır).
// Kullanıcı bilgisi düz alanlarla alınır; auth.User alınsaydı session→auth
// import döngüsü oluşurdu (auth handler'ları session'ı kullanıyor).
func (m *SessionManager) Create(ctx context.Context, c fiber.Ctx, scope Scope, userID, email, name, role string) (*Session, error) {
	sess := &Session{
		ID:        randomToken(),
		UserID:    userID,
		Email:     email,
		Name:      name,
		Role:      role,
		CSRFToken: randomToken(),
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	if err := m.save(ctx, scope, sess); err != nil {
		return nil, err
	}

	c.Cookie(&fiber.Cookie{
		Name:     scope.CookieName,
		Value:    sess.ID,
		Path:     scope.CookiePath,
		HTTPOnly: true,
		Secure:   m.secure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})

	return sess, nil
}

func (m *SessionManager) save(ctx context.Context, scope Scope, sess *Session) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}

	// Redis TTL = idle timeout; absolute timeout Get sırasında kontrol edilir.
	return m.rdb.Set(ctx, scope.RedisPrefix+sess.ID, data, m.idleTTL).Err()
}

// Get, cookie'den session'ı okur; idle + absolute timeout kontrolü yapar,
// geçerliyse LastSeen'i yenileyerek TTL'i kaydırır (sliding expiration).
func (m *SessionManager) Get(ctx context.Context, c fiber.Ctx, scope Scope) (*Session, error) {
	id := c.Cookies(scope.CookieName)
	if id == "" {
		return nil, ErrNoSession
	}

	data, err := m.rdb.Get(ctx, scope.RedisPrefix+id).Bytes()
	if err != nil {
		return nil, ErrNoSession
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, ErrNoSession
	}
	sess.ID = id

	now := time.Now()
	if now.Sub(sess.CreatedAt) > m.absoluteTTL || now.Sub(sess.LastSeen) > m.idleTTL {
		_ = m.rdb.Del(ctx, scope.RedisPrefix+id).Err()
		return nil, ErrNoSession
	}
	sess.LastSeen = now

	_ = m.save(ctx, scope, &sess)
	return &sess, nil
}

// Destroy, session'ı Redis'ten siler ve cookie'yi geçersizleştirir.
func (m *SessionManager) Destroy(ctx context.Context, c fiber.Ctx, scope Scope) {
	if id := c.Cookies(scope.CookieName); id != "" {
		_ = m.rdb.Del(ctx, scope.RedisPrefix+id).Err()
	}

	c.Cookie(&fiber.Cookie{
		Name:     scope.CookieName,
		Value:    "",
		Path:     scope.CookiePath,
		HTTPOnly: true,
		Secure:   m.secure,
		SameSite: fiber.CookieSameSiteLaxMode,
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
	})
}

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err) // crypto/rand hatası kurtarılamaz
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
