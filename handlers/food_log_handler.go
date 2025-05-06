package handlers

import (
	"context"
	"nitri-meal-backend/database"
	"nitri-meal-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetFoodLogsByUserID retrieves food logs for a specific user
func GetFoodLogsByUserID(c *fiber.Ctx) error {
    userID := c.Params("userId")
    if userID == "" {
        return c.Status(400).JSON(fiber.Map{
            "error": "User ID is required",
        })
    }

    collection := database.GetCollection("food_logs")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var foodLogs []models.FoodLog
    cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch food logs",
        })
    }
    defer cursor.Close(ctx)

    if err := cursor.All(ctx, &foodLogs); err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to decode food logs",
        })
    }

    return c.JSON(foodLogs)
}

// CreateFoodLog creates a new food log entry
func CreateFoodLog(c *fiber.Ctx) error {
    foodLog := new(models.FoodLog)
    if err := c.BodyParser(foodLog); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    collection := database.GetCollection("food_logs")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Get next log ID
    nextID, err := getNextFoodLogID(ctx)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to generate log ID",
        })
    }

    // Set the log ID and other fields
    foodLog.ID = primitive.NewObjectID()
    foodLog.LogID = nextID
    foodLog.CreatedAt = time.Now()

    // Save to database
    _, err = collection.InsertOne(ctx, foodLog)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to create food log",
        })
    }

    return c.Status(201).JSON(foodLog)
}

// Helper function to get next food log ID
func getNextFoodLogID(ctx context.Context) (int, error) {
    collection := database.GetCollection("food_logs")
    
    // Find the food log with highest log_id
    opts := options.FindOne().SetSort(bson.M{"id": -1})
    var lastLog models.FoodLog
    
    err := collection.FindOne(ctx, bson.M{}, opts).Decode(&lastLog)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return 1, nil // Start with ID 1 if no logs exist
        }
        return 0, err
    }
    
    return lastLog.LogID + 1, nil
}