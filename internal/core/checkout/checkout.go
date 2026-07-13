// Package checkout, sepet -> ödeme -> sipariş akışını orkestre eder.
package checkout

import (
	"context"
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"ecommerce/internal/core/cart"
	"ecommerce/internal/core/order"
	"ecommerce/internal/core/payment"
	"ecommerce/internal/core/product"
)

var ErrEmptyCart = errors.New("sepet boş")

type Service struct {
	carts    *cart.Service
	orders   *order.Service
	products *product.Service
	provider payment.Provider
}

func NewService(carts *cart.Service, orders *order.Service, products *product.Service, provider payment.Provider) *Service {
	return &Service{carts: carts, orders: orders, products: products, provider: provider}
}

type PlaceOrderInput struct {
	CartID   string
	UserID   string // misafirse boş
	Email    string
	FullName string
	Address  string
}

// PlaceOrder, sepetten pending_payment durumunda sipariş oluşturur ve
// ödeme başlatır. Dönen Payment'taki ThreeDSURL'e kullanıcı yönlendirilir.
func (s *Service) PlaceOrder(ctx context.Context, in PlaceOrderInput) (*order.Order, *payment.Payment, error) {
	if strings.TrimSpace(in.Email) == "" || strings.TrimSpace(in.FullName) == "" || strings.TrimSpace(in.Address) == "" {
		return nil, nil, errors.New("e-posta, ad soyad ve adres zorunlu")
	}
	c, err := s.carts.Get(ctx, in.CartID)
	if err != nil {
		return nil, nil, err
	}
	if len(c.Items) == 0 {
		return nil, nil, ErrEmptyCart
	}

	o := &order.Order{
		UserID:     in.UserID,
		Email:      strings.TrimSpace(in.Email),
		FullName:   strings.TrimSpace(in.FullName),
		Address:    strings.TrimSpace(in.Address),
		TotalCents: c.TotalCents,
		Status:     order.StatusPendingPayment,
	}
	for _, it := range c.Items {
		o.Items = append(o.Items, order.Item{
			ProductID:  it.Product.ID.Hex(),
			Name:       it.Product.Name,
			PriceCents: it.Product.PriceCents,
			Qty:        it.Qty,
		})
	}
	if err := s.orders.Create(ctx, o); err != nil {
		return nil, nil, err
	}

	pay, err := s.provider.CreatePayment(ctx, payment.CreateRequest{
		OrderID:     o.ID.Hex(),
		AmountCents: o.TotalCents,
		Currency:    "TRY",
		Email:       o.Email,
	})
	if err != nil {
		return nil, nil, err
	}
	o.PaymentID = pay.ID
	// payment_id'yi siparişe işle
	if err := s.orders.UpdatePaymentID(ctx, o.ID.Hex(), pay.ID); err != nil {
		return nil, nil, err
	}
	return o, pay, nil
}

// PollResult, htmx polling endpoint'inin kullandığı sonuç.
type PollResult struct {
	Done      bool
	Succeeded bool
	OrderID   string
}

// CheckPayment, provider'dan durumu sorar; sonuçlanmışsa siparişi günceller,
// başarılıysa stok düşer ve sepeti temizler.
func (s *Service) CheckPayment(ctx context.Context, orderID, cartID string) (*PollResult, error) {
	o, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	// idempotent: sipariş zaten sonuçlandıysa provider'a sormaya gerek yok
	switch o.Status {
	case order.StatusPaid:
		return &PollResult{Done: true, Succeeded: true, OrderID: orderID}, nil
	case order.StatusCancelled:
		return &PollResult{Done: true, Succeeded: false, OrderID: orderID}, nil
	}

	st, err := s.provider.GetStatus(ctx, o.PaymentID)
	if err != nil {
		return nil, err
	}
	switch st {
	case payment.StatusSucceeded:
		for _, it := range o.Items {
			oid, err := primitive.ObjectIDFromHex(it.ProductID)
			if err != nil {
				continue
			}
			_ = s.products.DecrementStock(ctx, oid, it.Qty) // stok eksiye düşmesin diye repo kontrol ediyor
		}
		if err := s.orders.UpdateStatus(ctx, orderID, order.StatusPaid); err != nil {
			return nil, err
		}
		_ = s.carts.Clear(ctx, cartID)
		return &PollResult{Done: true, Succeeded: true, OrderID: orderID}, nil
	case payment.StatusFailed:
		if err := s.orders.UpdateStatus(ctx, orderID, order.StatusCancelled); err != nil {
			return nil, err
		}
		return &PollResult{Done: true, Succeeded: false, OrderID: orderID}, nil
	default:
		return &PollResult{Done: false, OrderID: orderID}, nil
	}
}
