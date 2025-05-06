package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Meals struct {
    Breakfast string `json:"Breakfast" bson:"Breakfast"`
    Lunch     string `json:"Lunch" bson:"Lunch"`
    Dinner    string `json:"Dinner" bson:"Dinner"`
}

type MealPlan struct {
    ID      primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
    PlanID  int                 `json:"id" bson:"id"`
    UserID  string              `json:"user_id" bson:"user_id"`
    Date    string              `json:"date" bson:"date"`
    Meal    Meals               `json:"meal" bson:"meal"`
    Recipes []int               `json:"recipes" bson:"recipes"`
}