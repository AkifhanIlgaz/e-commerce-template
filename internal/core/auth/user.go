package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const (
	RoleAdmin    = "admin"
	RoleCustomer = "customer"
)

var (
	ErrInvalidCredentials = errors.New("e-posta veya şifre hatalı")
	ErrEmailTaken         = errors.New("bu e-posta ile kayıtlı kullanıcı var")
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Email        string             `bson:"email"`
	PasswordHash string             `bson:"password_hash"`
	Name         string             `bson:"name"`
	Role         string             `bson:"role"` // "admin" | "customer"
	CreatedAt    time.Time          `bson:"created_at"`
}

type UserRepository struct {
	col *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	col := db.Collection("users")
	_, _ = col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return &UserRepository{col: col}
}

// Service, kullanıcı kayıt/giriş iş mantığı. Session yönetimi SessionManager'da.
type Service struct {
	users *UserRepository
}

func NewService(users *UserRepository) *Service {
	return &Service{users: users}
}

func (s *Service) Register(ctx context.Context, email, password, name, role string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.New("geçerli bir e-posta girin")
	}
	if len(password) < 6 {
		return nil, errors.New("şifre en az 6 karakter olmalı")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &User{
		ID:           primitive.NewObjectID(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         strings.TrimSpace(name),
		Role:         role,
		CreatedAt:    time.Now(),
	}
	if _, err := s.users.col.InsertOne(ctx, u); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return u, nil
}

// Authenticate, email+şifre doğrular; requiredRole boş değilse rol de eşleşmeli.
// Admin girişi ile müşteri girişinin karışmaması bu rol kontrolüyle sağlanır.
func (s *Service) Authenticate(ctx context.Context, email, password, requiredRole string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var u User
	err := s.users.col.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		// timing farkını azaltmak için yine de bir bcrypt çalıştır
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$invalidinvalidinvalidinvalidinvalidinvalid"), []byte(password))
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	if requiredRole != "" && u.Role != requiredRole {
		return nil, ErrInvalidCredentials
	}
	return &u, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("kullanıcı bulunamadı")
	}
	var u User
	if err := s.users.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&u); err != nil {
		return nil, errors.New("kullanıcı bulunamadı")
	}
	return &u, nil
}

func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	cur, err := s.users.col.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	var out []User
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EnsureAdmin, ilk açılışta env'den gelen admin kullanıcısını oluşturur (varsa dokunmaz).
func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	err := s.users.col.FindOne(ctx, bson.M{"email": strings.ToLower(email), "role": RoleAdmin}).Err()
	if err == nil {
		return nil // zaten var
	}
	_, err = s.Register(ctx, email, password, "Admin", RoleAdmin)
	if errors.Is(err, ErrEmailTaken) {
		return nil
	}
	return err
}
