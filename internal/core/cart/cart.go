// Package cart, sepeti Redis'te tutar. Sepet, session'dan bağımsız bir
// "cart_id" cookie'sine bağlıdır — misafir kullanıcı da sepet kullanabilir.
package cart

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"ecommerce/internal/core/product"
)

const (
	cookieName  = "cart_id"
	redisPrefix = "cart:"
	cartTTL     = 30 * 24 * time.Hour
)

type Item struct {
	ProductID string `json:"product_id"`
	Qty       int    `json:"qty"`
}

// Cart, ürün bilgileriyle zenginleştirilmiş sepet görünümü.
type Cart struct {
	Items      []CartItem
	TotalCents int64
}

type CartItem struct {
	Product product.Product
	Qty     int
}

func (c Cart) TotalDisplay() string { return product.FormatPrice(c.TotalCents) }

func (c Cart) Count() int {
	n := 0
	for _, it := range c.Items {
		n += it.Qty
	}
	return n
}

type Service struct {
	rdb      *redis.Client
	products *product.Service
	secure   bool
}

func NewService(rdb *redis.Client, products *product.Service, secure bool) *Service {
	return &Service{rdb: rdb, products: products, secure: secure}
}

// CartID, cookie'den sepet ID'sini okur; yoksa oluşturup cookie yazar.
func (s *Service) CartID(c fiber.Ctx) string {
	if v := c.Cookies(cookieName); v != "" {
		return v
	}
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	id := base64.RawURLEncoding.EncodeToString(b)
	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    id,
		Path:     "/",
		HTTPOnly: true,
		Secure:   s.secure,
		SameSite: fiber.CookieSameSiteLaxMode,
		MaxAge:   int(cartTTL.Seconds()),
	})
	return id
}

func (s *Service) rawItems(ctx context.Context, cartID string) ([]Item, error) {
	data, err := s.rdb.Get(ctx, redisPrefix+cartID).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Service) saveItems(ctx context.Context, cartID string, items []Item) error {
	if len(items) == 0 {
		return s.rdb.Del(ctx, redisPrefix+cartID).Err()
	}
	data, _ := json.Marshal(items)
	return s.rdb.Set(ctx, redisPrefix+cartID, data, cartTTL).Err()
}

func (s *Service) Add(ctx context.Context, cartID, productID string, qty int) error {
	if qty < 1 {
		qty = 1
	}
	// ürün var mı ve aktif mi kontrol et
	p, err := s.products.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if !p.Active || p.Stock <= 0 {
		return errors.New("ürün satışta değil")
	}
	items, err := s.rawItems(ctx, cartID)
	if err != nil {
		return err
	}
	for i := range items {
		if items[i].ProductID == productID {
			items[i].Qty += qty
			return s.saveItems(ctx, cartID, items)
		}
	}
	items = append(items, Item{ProductID: productID, Qty: qty})
	return s.saveItems(ctx, cartID, items)
}

// SetQty, qty <= 0 ise ürünü sepetten çıkarır.
func (s *Service) SetQty(ctx context.Context, cartID, productID string, qty int) error {
	items, err := s.rawItems(ctx, cartID)
	if err != nil {
		return err
	}
	out := items[:0]
	for _, it := range items {
		if it.ProductID == productID {
			if qty <= 0 {
				continue
			}
			it.Qty = qty
		}
		out = append(out, it)
	}
	return s.saveItems(ctx, cartID, out)
}

func (s *Service) Remove(ctx context.Context, cartID, productID string) error {
	return s.SetQty(ctx, cartID, productID, 0)
}

func (s *Service) Clear(ctx context.Context, cartID string) error {
	return s.rdb.Del(ctx, redisPrefix+cartID).Err()
}

// Get, sepeti ürün bilgileriyle birlikte döner. Silinen/pasif ürünler atlanır.
func (s *Service) Get(ctx context.Context, cartID string) (*Cart, error) {
	items, err := s.rawItems(ctx, cartID)
	if err != nil {
		return nil, err
	}
	cart := &Cart{}
	for _, it := range items {
		p, err := s.products.GetByID(ctx, it.ProductID)
		if err != nil || !p.Active {
			continue
		}
		cart.Items = append(cart.Items, CartItem{Product: *p, Qty: it.Qty})
		cart.TotalCents += p.PriceCents * int64(it.Qty)
	}
	return cart, nil
}
