package tag

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Tag struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Name      string        `bson:"name"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}
