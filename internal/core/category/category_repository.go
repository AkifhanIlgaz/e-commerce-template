package category

import (
	"context"
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type categoryRepository struct {
	col *mongo.Collection
}

func NewCategoryRepository(db *mongo.Database) *categoryRepository {
	col := db.Collection("categories")

	_, _ = col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &categoryRepository{col: col}
}

func (r *categoryRepository) Insert(ctx context.Context, b *Category) error {
	if _, err := r.col.InsertOne(ctx, b); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrCategoryNameTaken
		}
		return err
	}

	return nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id bson.ObjectID) (*Category, error) {
	var b Category
	filter := bson.M{"_id": id}

	if err := r.col.FindOne(ctx, filter).Decode(&b); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	return &b, nil
}

// FindPage, sayfalı liste ve toplam kayıt sayısı döner. query doluysa
// isimde büyük/küçük harf duyarsız kısmi eşleşme aranır.
func (r *categoryRepository) FindPage(ctx context.Context, query string, page, perPage int) ([]Category, int64, error) {
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

	var out []Category
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

func (r *categoryRepository) Update(ctx context.Context, b *Category) error {
	filter := bson.M{"_id": b.ID}
	update := bson.M{"$set": bson.M{"name": b.Name, "updated_at": b.UpdatedAt}}

	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrCategoryNameTaken
		}
		return err
	}
	if res.MatchedCount == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
