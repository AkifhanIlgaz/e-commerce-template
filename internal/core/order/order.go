package order

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ecommerce/internal/core/product"
)

const (
	StatusPendingPayment = "pending_payment"
	StatusPaid           = "paid"
	StatusShipped        = "shipped"
	StatusDelivered      = "delivered"
	StatusCancelled      = "cancelled"
)

// AllStatuses, admin'deki durum dropdown'ı için.
var AllStatuses = []string{StatusPendingPayment, StatusPaid, StatusShipped, StatusDelivered, StatusCancelled}

var ErrNotFound = errors.New("sipariş bulunamadı")

type Item struct {
	ProductID  string `bson:"product_id"`
	Name       string `bson:"name"`        // sipariş anındaki isim (ürün sonradan değişse de sabit kalır)
	PriceCents int64  `bson:"price_cents"` // sipariş anındaki fiyat
	Qty        int    `bson:"qty"`
}

func (i Item) LineTotalDisplay() string { return product.FormatPrice(i.PriceCents * int64(i.Qty)) }
func (i Item) PriceDisplay() string     { return product.FormatPrice(i.PriceCents) }

type Order struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserID     string             `bson:"user_id,omitempty"` // misafir siparişte boş
	Email      string             `bson:"email"`
	FullName   string             `bson:"full_name"`
	Address    string             `bson:"address"`
	Items      []Item             `bson:"items"`
	TotalCents int64              `bson:"total_cents"`
	Status     string             `bson:"status"`
	PaymentID  string             `bson:"payment_id"`
	CreatedAt  time.Time          `bson:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at"`
}

func (o Order) TotalDisplay() string { return product.FormatPrice(o.TotalCents) }

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("orders")}
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, o *Order) error {
	o.ID = primitive.NewObjectID()
	o.CreatedAt = time.Now()
	o.UpdatedAt = o.CreatedAt
	_, err := s.repo.col.InsertOne(ctx, o)
	return err
}

func (s *Service) GetByID(ctx context.Context, id string) (*Order, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrNotFound
	}
	var o Order
	if err := s.repo.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&o); err != nil {
		return nil, ErrNotFound
	}
	return &o, nil
}

func (s *Service) List(ctx context.Context) ([]Order, error) {
	cur, err := s.repo.col.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	var out []Order
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) ListByUser(ctx context.Context, userID string) ([]Order, error) {
	cur, err := s.repo.col.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	var out []Order
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id, status string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrNotFound
	}
	valid := false
	for _, st := range AllStatuses {
		if st == status {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("geçersiz sipariş durumu")
	}
	res, err := s.repo.col.UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) UpdatePaymentID(ctx context.Context, id, paymentID string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrNotFound
	}
	_, err = s.repo.col.UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$set": bson.M{"payment_id": paymentID, "updated_at": time.Now()}})
	return err
}

// Stats, admin dashboard için basit sayılar.
type Stats struct {
	TotalOrders   int64
	PendingOrders int64
	RevenueCents  int64 // sadece paid ve sonrası
}

func (st Stats) RevenueDisplay() string { return product.FormatPrice(st.RevenueCents) }

func (s *Service) Stats(ctx context.Context) (*Stats, error) {
	total, err := s.repo.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	pending, err := s.repo.col.CountDocuments(ctx, bson.M{"status": StatusPendingPayment})
	if err != nil {
		return nil, err
	}
	cur, err := s.repo.col.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"status": bson.M{"$in": []string{StatusPaid, StatusShipped, StatusDelivered}}}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "revenue": bson.M{"$sum": "$total_cents"}}}},
	})
	if err != nil {
		return nil, err
	}
	var agg []struct {
		Revenue int64 `bson:"revenue"`
	}
	if err := cur.All(ctx, &agg); err != nil {
		return nil, err
	}
	st := &Stats{TotalOrders: total, PendingOrders: pending}
	if len(agg) > 0 {
		st.RevenueCents = agg[0].Revenue
	}
	return st, nil
}
