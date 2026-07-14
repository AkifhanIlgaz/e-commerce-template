package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

// authService, kullanıcı kayıt/giriş iş mantığı. Session yönetimi SessionManager'da.
type authService struct {
	repo *authRepository
}

func NewAuthService(repo *authRepository) *authService {
	return &authService{repo: repo}
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &User{
		ID:           bson.NewObjectID(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         strings.TrimSpace(req.Name),
		Role:         req.Role,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.Insert(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Authenticate, email+şifre doğrular; RequiredRole boş değilse rol de eşleşmeli.
// Admin girişi ile müşteri girişinin karışmaması bu rol kontrolüyle sağlanır.
func (s *authService) Authenticate(ctx context.Context, req LoginRequest) (*User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		// timing farkını azaltmak için yine de bir bcrypt çalıştır
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$invalidinvalidinvalidinvalidinvalidinvalid"), []byte(req.Password))
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if req.RequiredRole != "" && u.Role != req.RequiredRole {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}

func (s *authService) GetByID(ctx context.Context, id string) (*User, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return s.repo.FindByID(ctx, oid)
}

func (s *authService) ListUsers(ctx context.Context) ([]User, error) {
	return s.repo.FindAll(ctx)
}

// EnsureAdmin, ilk açılışta env'den gelen admin kullanıcısını oluşturur (varsa dokunmaz).
func (s *authService) EnsureAdmin(ctx context.Context, email, password string) error {
	exists, err := s.repo.AdminExists(ctx, strings.ToLower(email))
	if err == nil && exists {
		return nil // zaten var
	}
	_, err = s.Register(ctx, RegisterRequest{Email: email, Password: password, Name: "Admin", Role: RoleAdmin})
	if errors.Is(err, ErrEmailTaken) {
		return nil
	}
	return err
}
