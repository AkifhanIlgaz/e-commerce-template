package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("E-posta veya şifre hatalı")
	ErrEmailTaken         = errors.New("Bu e-posta ile kayıtlı kullanıcı var")
	ErrUserNotFound       = errors.New("Kullanıcı bulunamadı")
)
