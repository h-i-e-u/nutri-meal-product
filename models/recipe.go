package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Ingredient struct {
    Name   string `json:"name" bson:"name"`
    Amount string `json:"amount" bson:"amount"`
}

type NutritionInfo struct {
    Calories int `json:"calories" bson:"calories"`
    Protein  int `json:"protein" bson:"protein"`
    Carbs    int `json:"carbs" bson:"carbs"`
    Fat      int `json:"fat" bson:"fat"`
}

type Recipe struct {
    ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    RecipeID        int               `json:"id" bson:"id"`
    Name            string            `json:"name" bson:"name"`
    Ingredients     []Ingredient      `json:"ingredients" bson:"ingredients"`
    Instructions    string            `json:"instructions" bson:"instructions"`
    NutritionInfo   NutritionInfo    `json:"nutrition_info" bson:"nutrition_info"`
    Category        string            `json:"category" bson:"category"`
    Tips            string            `json:"tips" bson:"tips"`
    PreparationTime string            `json:"preparationTime" bson:"preparation_time"`
    Difficulty      string            `json:"difficulty" bson:"difficulty"`
    Allergens       []string          `json:"allergens" bson:"allergens"`
}