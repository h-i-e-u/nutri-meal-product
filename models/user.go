package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Email      string            `json:"email" bson:"email"`
	Name       string            `json:"name" bson:"name"`
	Picture    string            `json:"picture" bson:"picture"`
	Height     float64           `json:"height" bson:"height"`
	Birthday   string            `json:"birthday,omitempty" bson:"birthday,omitempty"`
	CreatedAt  time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at" bson:"updated_at"`
	DeleteHash string            `json:"deleteHash,omitempty" bson:"deleteHash,omitempty"`
}
