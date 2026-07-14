package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	RoleAdmin    = "admin"
	RoleCustomer = "customer"
)

type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	Email        string        `bson:"email"`
	PasswordHash string        `bson:"password_hash"`
	Name         string        `bson:"name"`
	Role         string        `bson:"role"` // "admin" | "customer"
	CreatedAt    time.Time     `bson:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"`
}
