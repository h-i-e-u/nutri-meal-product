package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Author struct {
    ID      string `bson:"_id" json:"_id"`
    Name    string `bson:"name" json:"name"`
    Picture string `bson:"picture" json:"picture"`
}

type Post struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
    Content   string            `bson:"content" json:"content"`
    Image     []byte            `bson:"image,omitempty" json:"image,omitempty"`
    ImageType string            `bson:"imageType,omitempty" json:"imageType,omitempty"` // Store MIME type
    CreatedAt time.Time         `bson:"createdAt" json:"createdAt"`
    Likes     int               `bson:"likes" json:"likes"`
    LikedBy   []string          `bson:"likedBy" json:"likedBy"`
    Author    Author            `bson:"author" json:"author"`
}