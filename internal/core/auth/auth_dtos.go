package auth

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// friendlyValidation, validator'ın ham (İngilizce, alan adlı) hatasını
// kullanıcıya gösterilebilir Türkçe mesaja çevirir. Handler'lar Validate()
// çıktısını doğrudan forma bastığı için çeviri burada yapılır.
func friendlyValidation(err error) error {
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) || len(verrs) == 0 {
		return err
	}
	fe := verrs[0]
	switch fe.Field() {
	case "Email":
		if fe.Tag() == "email" {
			return errors.New("Geçerli bir e-posta adresi girin")
		}
		return errors.New("E-posta zorunludur")
	case "Password":
		if fe.Tag() == "min" {
			return errors.New("Şifre en az " + fe.Param() + " karakter olmalıdır")
		}
		return errors.New("Şifre zorunludur")
	case "Name":
		return errors.New("İsim en fazla " + fe.Param() + " karakter olabilir")
	}
	return errors.New("Form bilgileri geçersiz")
}

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
	return friendlyValidation(validate.Struct(r))
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
	return friendlyValidation(validate.Struct(r))
}
