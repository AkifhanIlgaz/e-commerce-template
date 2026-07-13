package product

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Slug        string             `bson:"slug"`
	Description string             `bson:"description"`
	PriceCents  int64              `bson:"price_cents"` // kuruş cinsinden — float kullanma
	Currency    string             `bson:"currency"`    // "TRY"
	ImageURL    string             `bson:"image_url"`
	Stock       int                `bson:"stock"`
	Active      bool               `bson:"active"` // false ise storefront'ta görünmez
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

// PriceDisplay, kuruşu "199,90" formatında döner (templ'lerde kullanılır).
func (p Product) PriceDisplay() string {
	return FormatPrice(p.PriceCents)
}
