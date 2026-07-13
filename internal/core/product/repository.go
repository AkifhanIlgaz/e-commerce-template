package product

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	col := db.Collection("products")
	// slug unique index
	_, _ = col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return &Repository{col: col}
}

func (r *Repository) Create(ctx context.Context, p *Product) error {
	p.ID = primitive.NewObjectID()
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	_, err := r.col.InsertOne(ctx, p)
	return err
}

func (r *Repository) Update(ctx context.Context, p *Product) error {
	p.UpdatedAt = time.Now()
	_, err := r.col.ReplaceOne(ctx, bson.M{"_id": p.ID}, p)
	return err
}

func (r *Repository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *Repository) GetByID(ctx context.Context, id primitive.ObjectID) (*Product, error) {
	var p Product
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (*Product, error) {
	var p Product
	err := r.col.FindOne(ctx, bson.M{"slug": slug}).Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type ListFilter struct {
	Search     string // isim içinde arama (case-insensitive)
	ActiveOnly bool   // storefront true, admin false kullanır
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]Product, error) {
	q := bson.M{}
	if f.ActiveOnly {
		q["active"] = true
	}
	if f.Search != "" {
		q["name"] = bson.M{"$regex": f.Search, "$options": "i"}
	}
	cur, err := r.col.Find(ctx, q, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	var out []Product
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DecrementStock, stok yeterliyse azaltır; yetersizse hata döner.
func (r *Repository) DecrementStock(ctx context.Context, id primitive.ObjectID, qty int) error {
	res, err := r.col.UpdateOne(ctx,
		bson.M{"_id": id, "stock": bson.M{"$gte": qty}},
		bson.M{"$inc": bson.M{"stock": -qty}, "$set": bson.M{"updated_at": time.Now()}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrInsufficientStock
	}
	return nil
}

func (r *Repository) Count(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{})
}
