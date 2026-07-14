package brand

import (
	"context"
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type brandRepository struct {
	col *mongo.Collection
}

func NewBrandRepository(db *mongo.Database) *brandRepository {
	col := db.Collection("brands")

	_, _ = col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &brandRepository{col: col}
}

func (r *brandRepository) Insert(ctx context.Context, b *Brand) error {
	if _, err := r.col.InsertOne(ctx, b); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrBrandNameTaken
		}
		return err
	}

	return nil
}

func (r *brandRepository) FindByID(ctx context.Context, id bson.ObjectID) (*Brand, error) {
	var b Brand
	filter := bson.M{"_id": id}

	if err := r.col.FindOne(ctx, filter).Decode(&b); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrBrandNotFound
		}
		return nil, err
	}

	return &b, nil
}

// FindPage, sayfalı liste ve toplam kayıt sayısı döner. query doluysa
// isimde büyük/küçük harf duyarsız kısmi eşleşme aranır.
func (r *brandRepository) FindPage(ctx context.Context, query string, page, perPage int) ([]Brand, int64, error) {
	filter := bson.M{}
	if query != "" {
		filter["name"] = bson.M{"$regex": regexp.QuoteMeta(query), "$options": "i"}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64(page-1) * int64(perPage)).
		SetLimit(int64(perPage))

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	var out []Brand
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

func (r *brandRepository) Update(ctx context.Context, b *Brand) error {
	filter := bson.M{"_id": b.ID}
	update := bson.M{"$set": bson.M{"name": b.Name, "updated_at": b.UpdatedAt}}

	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrBrandNameTaken
		}
		return err
	}
	if res.MatchedCount == 0 {
		return ErrBrandNotFound
	}

	return nil
}

func (r *brandRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrBrandNotFound
	}

	return nil
}
