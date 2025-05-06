package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FoodLog struct {
    ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    LogID     int               `json:"id" bson:"id"`
    UserID    string            `json:"user_id" bson:"user_id"`
    FoodName  string            `json:"food_name" bson:"food_name"`
    FoodID    int              `json:"food_id" bson:"food_id"`
    Calories  int               `json:"calories" bson:"calories"`
    Protein   int               `json:"protein" bson:"protein"`
    Carbs     int              `json:"carbs" bson:"carbs"`
    Fat       int              `json:"fat" bson:"fat"`
    MealTime  string           `json:"meal_time" bson:"meal_time"`
    Date      string           `json:"date" bson:"date"`
    CreatedAt time.Time        `json:"created_at" bson:"created_at"`
}