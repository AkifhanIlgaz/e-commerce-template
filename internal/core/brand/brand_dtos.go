package brand

import (
	"errors"
	"strings"

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
	if fe.Field() == "Name" {
		if fe.Tag() == "max" {
			return errors.New("İsim en fazla " + fe.Param() + " karakter olabilir")
		}
		return errors.New("İsim zorunludur")
	}
	return errors.New("Form bilgileri geçersiz")
}

// CreateBrandRequest, Create servisinin girdisi. Handler formdan doldurur;
// servis içinde Brand modeline çevrilir.
type CreateBrandRequest struct {
	Name string `json:"name" form:"name" validate:"required,max=100"`
}

func (r CreateBrandRequest) Validate() error {
	return friendlyValidation(validate.Struct(r))
}

// UpdateBrandRequest, Update servisinin girdisi. ID handler tarafında
// path parametresinden set edilir.
type UpdateBrandRequest struct {
	ID   string `json:"-" form:"-" validate:"required"`
	Name string `json:"name" form:"name" validate:"required,max=100"`
}

func (r UpdateBrandRequest) Validate() error {
	return friendlyValidation(validate.Struct(r))
}

// ListBrandsRequest, sayfalı listeleme girdisi. Page/PerPage sınır
// dışıysa Normalize varsayılana çeker.
type ListBrandsRequest struct {
	Page    int    `json:"page" form:"page"`
	PerPage int    `json:"per_page" form:"per_page"`
	Query   string `json:"q" form:"q"`
}

const (
	defaultPerPage = 20
	maxPerPage     = 100
)

// Normalize, page/per_page değerlerini güvenli aralığa çeker.
func (r ListBrandsRequest) Normalize() ListBrandsRequest {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PerPage < 1 {
		r.PerPage = defaultPerPage
	}
	if r.PerPage > maxPerPage {
		r.PerPage = maxPerPage
	}
	r.Query = strings.TrimSpace(r.Query)
	return r
}

// BrandPage, sayfalı listeleme çıktısı.
type BrandPage struct {
	Items      []Brand
	Total      int64
	Page       int
	PerPage    int
	TotalPages int
}
