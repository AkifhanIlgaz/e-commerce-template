package category

import "errors"

var (
	ErrCategoryNotFound  = errors.New("Kategori bulunamadı")
	ErrCategoryNameTaken = errors.New("Bu isimde bir kategori zaten var")
	ErrInvalidID         = errors.New("Geçersiz kimlik")
)
