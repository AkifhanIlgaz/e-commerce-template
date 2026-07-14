package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("e-posta veya şifre hatalı")
	ErrEmailTaken         = errors.New("bu e-posta ile kayıtlı kullanıcı var")
	ErrUserNotFound       = errors.New("kullanıcı bulunamadı")
)
