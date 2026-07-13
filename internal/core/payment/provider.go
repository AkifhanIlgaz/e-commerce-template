// Package payment, ödeme sağlayıcısını soyutlar. Yeni müşteride gerçek bir
// sağlayıcı (iyzico, stripe, ...) eklemek için Provider interface'ini
// implemente edip factory'ye (NewProvider) bir case eklemek yeterli —
// checkout akışı ve handler'lar değişmez.
package payment

import (
	"context"
	"fmt"

	"ecommerce/internal/config"
)

type Status string

const (
	StatusPending   Status = "pending" // 3DS onayı bekleniyor
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

type CreateRequest struct {
	OrderID     string
	AmountCents int64
	Currency    string
	Email       string
}

type Payment struct {
	ID          string
	Status      Status
	ThreeDSURL  string // kullanıcının onay vereceği sayfa (iframe'de açılır)
	AmountCents int64
}

type Provider interface {
	Name() string
	// CreatePayment, ödeme başlatır. 3DS gerekiyorsa Status=pending ve
	// ThreeDSURL dolu döner; kullanıcı orada onayladıktan sonra
	// GetStatus polling ile sonuç alınır.
	CreatePayment(ctx context.Context, req CreateRequest) (*Payment, error)
	GetStatus(ctx context.Context, paymentID string) (Status, error)
}

// NewProvider, config'e göre sağlayıcı seçer. API anahtarları env'den gelir.
func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.PaymentProvider {
	case "mock", "":
		return NewMockProvider(), nil
	// case "iyzico":
	//     return NewIyzicoProvider(cfg.PaymentAPIKey, cfg.PaymentSecretKey, cfg.PaymentBaseURL), nil
	default:
		return nil, fmt.Errorf("bilinmeyen payment provider: %q", cfg.PaymentProvider)
	}
}
