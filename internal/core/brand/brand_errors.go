package brand

import "errors"

var (
	ErrBrandNotFound  = errors.New("Marka bulunamadı")
	ErrBrandNameTaken = errors.New("Bu isimde bir marka zaten var")
	ErrInvalidID      = errors.New("Geçersiz kimlik")
)
