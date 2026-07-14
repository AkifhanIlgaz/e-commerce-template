package auth

import "github.com/go-playground/validator/v10"

var validate = validator.New()

// RegisterRequest, Register servisinin girdisi. Handler formdan doldurur;
// servis içinde User modeline çevrilir. Role client'tan alınmaz (mass
// assignment önlemi), handler tarafında server-side set edilir.
type RegisterRequest struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required,min=6"`
	Name     string `json:"name" form:"name" validate:"max=100"`
	Role     string `json:"-" form:"-" validate:"required,oneof=admin customer"`
}

func (r RegisterRequest) Validate() error {
	return validate.Struct(r)
}

// LoginRequest, Authenticate servisinin girdisi. RequiredRole boş değilse
// kullanıcının rolü de eşleşmek zorundadır (admin/müşteri girişi ayrımı);
// client'tan alınmaz, handler tarafında set edilir.
type LoginRequest struct {
	Email        string `json:"email" form:"email" validate:"required"`
	Password     string `json:"password" form:"password" validate:"required"`
	RequiredRole string `json:"-" form:"-" validate:"omitempty,oneof=admin customer"`
}

func (r LoginRequest) Validate() error {
	return validate.Struct(r)
}
