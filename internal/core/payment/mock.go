package payment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

// MockProvider, gerçek bir 3DS akışını simüle eder:
//  1. CreatePayment -> pending + /pay/mock-3ds?payment_id=... URL'i
//  2. Kullanıcı o sayfada (iframe) "Onayla" ya da "Reddet" der
//  3. Storefront polling ile GetStatus sorar, sonuç succeeded/failed olur
//
// Ödemeler bellekte tutulur — mock olduğu için restart'ta kaybolması sorun değil.
type MockProvider struct {
	mu       sync.Mutex
	payments map[string]*Payment
}

func NewMockProvider() *MockProvider {
	return &MockProvider{payments: make(map[string]*Payment)}
}

func (m *MockProvider) Name() string { return "mock" }

func (m *MockProvider) CreatePayment(_ context.Context, req CreateRequest) (*Payment, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	id := "mockpay_" + hex.EncodeToString(b)
	p := &Payment{
		ID:          id,
		Status:      StatusPending,
		ThreeDSURL:  "/pay/mock-3ds?payment_id=" + id,
		AmountCents: req.AmountCents,
	}
	m.mu.Lock()
	m.payments[id] = p
	m.mu.Unlock()
	return p, nil
}

func (m *MockProvider) GetStatus(_ context.Context, paymentID string) (Status, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.payments[paymentID]
	if !ok {
		return "", errors.New("ödeme bulunamadı")
	}
	return p.Status, nil
}

// Resolve, mock 3DS sayfasındaki Onayla/Reddet butonlarının çağırdığı metod.
func (m *MockProvider) Resolve(paymentID string, approved bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.payments[paymentID]
	if !ok {
		return errors.New("ödeme bulunamadı")
	}
	if p.Status != StatusPending {
		return errors.New("ödeme zaten sonuçlanmış")
	}
	if approved {
		p.Status = StatusSucceeded
	} else {
		p.Status = StatusFailed
	}
	return nil
}

func (m *MockProvider) Amount(paymentID string) (int64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.payments[paymentID]
	if !ok {
		return 0, false
	}
	return p.AmountCents, true
}
