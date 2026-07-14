package brand

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// brandService, marka CRUD iş mantığı. DTO'ları entity'ye çevirmek ve
// ID doğrulaması bu katmandadır.
type brandService struct {
	repo *brandRepository
}

func NewBrandService(repo *brandRepository) *brandService {
	return &brandService{repo: repo}
}

func (s *brandService) Create(ctx context.Context, req CreateBrandRequest) (*Brand, error) {
	b := &Brand{
		ID:        bson.NewObjectID(),
		Name:      strings.TrimSpace(req.Name),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Insert(ctx, b); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *brandService) Update(ctx context.Context, req UpdateBrandRequest) (*Brand, error) {
	oid, err := bson.ObjectIDFromHex(req.ID)
	if err != nil {
		return nil, ErrInvalidID
	}

	b, err := s.repo.FindByID(ctx, oid)
	if err != nil {
		return nil, err
	}

	b.Name = strings.TrimSpace(req.Name)
	b.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, b); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *brandService) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}
	return s.repo.Delete(ctx, oid)
}

func (s *brandService) GetByID(ctx context.Context, id string) (*Brand, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}
	return s.repo.FindByID(ctx, oid)
}

// List, sayfalı liste döner.
func (s *brandService) List(ctx context.Context, req ListBrandsRequest) (*BrandPage, error) {
	req = req.Normalize()

	items, total, err := s.repo.FindPage(ctx, req.Query, req.Page, req.PerPage)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.PerPage) - 1) / int64(req.PerPage))

	return &BrandPage{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalPages: totalPages,
	}, nil
}
