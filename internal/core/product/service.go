package product

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInsufficientStock = errors.New("yetersiz stok")
	ErrNotFound          = errors.New("ürün bulunamadı")
)

// Service, hem admin hem storefront tarafından kullanılan ürün iş mantığı.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	Name        string
	Description string
	PriceCents  int64
	ImageURL    string
	Stock       int
	Active      bool
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*Product, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.New("ürün adı boş olamaz")
	}
	if in.PriceCents <= 0 {
		return nil, errors.New("fiyat sıfırdan büyük olmalı")
	}
	p := &Product{
		Name:        strings.TrimSpace(in.Name),
		Slug:        Slugify(in.Name),
		Description: in.Description,
		PriceCents:  in.PriceCents,
		Currency:    "TRY",
		ImageURL:    in.ImageURL,
		Stock:       in.Stock,
		Active:      in.Active,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		// slug çakışırsa sonuna kısa bir ek koy
		p.Slug = p.Slug + "-" + p.ID.Hex()[18:]
		if err2 := s.repo.Create(ctx, p); err2 != nil {
			return nil, err2
		}
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id string, in CreateInput) (*Product, error) {
	p, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Name = strings.TrimSpace(in.Name)
	p.Description = in.Description
	p.PriceCents = in.PriceCents
	p.ImageURL = in.ImageURL
	p.Stock = in.Stock
	p.Active = in.Active
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrNotFound
	}
	return s.repo.Delete(ctx, oid)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Product, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrNotFound
	}
	p, err := s.repo.GetByID(ctx, oid)
	if err != nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*Product, error) {
	p, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]Product, error) {
	return s.repo.List(ctx, f)
}

func (s *Service) DecrementStock(ctx context.Context, id primitive.ObjectID, qty int) error {
	return s.repo.DecrementStock(ctx, id, qty)
}

func (s *Service) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

// --- yardımcılar ---

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

var trReplacer = strings.NewReplacer(
	"ç", "c", "ğ", "g", "ı", "i", "ö", "o", "ş", "s", "ü", "u",
	"Ç", "c", "Ğ", "g", "İ", "i", "I", "i", "Ö", "o", "Ş", "s", "Ü", "u",
)

func Slugify(s string) string {
	s = trReplacer.Replace(s)
	s = strings.ToLower(s)
	s = slugRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// FormatPrice, kuruşu "1.234,56 ₺" formatında döner.
func FormatPrice(cents int64) string {
	lira := cents / 100
	kurus := cents % 100
	// binlik ayraç
	str := fmt.Sprintf("%d", lira)
	var b strings.Builder
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	return fmt.Sprintf("%s,%02d ₺", b.String(), kurus)
}
