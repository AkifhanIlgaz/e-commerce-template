package tag

import "errors"

var (
	ErrTagNotFound  = errors.New("Etiket bulunamadı")
	ErrTagNameTaken = errors.New("Bu isimde bir etiket zaten var")
	ErrInvalidID    = errors.New("Geçersiz kimlik")
)
