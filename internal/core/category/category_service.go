package category

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// categoryService, kategori CRUD iş mantığı. DTO'ları entity'ye çevirmek ve
// ID doğrulaması bu katmandadır.
type categoryService struct {
	repo *categoryRepository
}

func NewCategoryService(repo *categoryRepository) *categoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) Create(ctx context.Context, req CreateCategoryRequest) (*Category, error) {
	b := &Category{
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

func (s *categoryService) Update(ctx context.Context, req UpdateCategoryRequest) (*Category, error) {
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

func (s *categoryService) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}
	return s.repo.Delete(ctx, oid)
}

func (s *categoryService) GetByID(ctx context.Context, id string) (*Category, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}
	return s.repo.FindByID(ctx, oid)
}

// List, sayfalı liste döner.
func (s *categoryService) List(ctx context.Context, req ListCategoriesRequest) (*CategoryPage, error) {
	req = req.Normalize()

	items, total, err := s.repo.FindPage(ctx, req.Query, req.Page, req.PerPage)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(req.PerPage) - 1) / int64(req.PerPage))

	return &CategoryPage{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalPages: totalPages,
	}, nil
}
