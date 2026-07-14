package auth

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type authRepository struct {
	col *mongo.Collection
}

func NewAuthRepository(db *mongo.Database) *authRepository {
	col := db.Collection("users")
	_, _ = col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return &authRepository{col: col}
}

func (r *authRepository) Insert(ctx context.Context, u *User) error {
	if _, err := r.col.InsertOne(ctx, u); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrEmailTaken
		}
		return err
	}

	return nil
}

func (r *authRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	filter := bson.M{"email": email}

	if err := r.col.FindOne(ctx, filter).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *authRepository) FindByID(ctx context.Context, id bson.ObjectID) (*User, error) {
	var u User
	filter := bson.M{"_id": id}

	if err := r.col.FindOne(ctx, filter).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *authRepository) FindAll(ctx context.Context) ([]User, error) {
	filter := bson.M{}
	sort := bson.D{{Key: "created_at", Value: -1}}

	cur, err := r.col.Find(ctx, filter, options.Find().SetSort(sort))
	if err != nil {
		return nil, err
	}

	var out []User
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *authRepository) AdminExists(ctx context.Context, email string) (bool, error) {
	filter := bson.M{"email": email, "role": RoleAdmin}

	if err := r.col.FindOne(ctx, filter).Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
