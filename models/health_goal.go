package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HealthGoal struct {
    ID                 primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
    UserID             primitive.ObjectID `json:"user_id" bson:"user_id"`
    TargetWeight       float64           `json:"targetWeight" bson:"target_weight"`
    CurrentWeight      float64           `json:"currentWeight" bson:"current_weight"`
    ActivityLevel      string            `json:"activityLevel" bson:"activity_level"`
    DietaryPreferences []string          `json:"dietaryPreferences" bson:"dietary_preferences"`
    WeeklyGoal         string            `json:"weeklyGoal" bson:"weekly_goal"`
    CreatedAt          time.Time         `json:"createdAt" bson:"created_at"`
    UpdatedAt          time.Time         `json:"updatedAt" bson:"updated_at"`
}