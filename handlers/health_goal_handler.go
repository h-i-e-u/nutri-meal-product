package handlers

import (
	"context"
	"log"
	"nitri-meal-backend/database"
	"nitri-meal-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateHealthGoal creates a new health goal
func CreateHealthGoal(c *fiber.Ctx) error {
    goal := new(models.HealthGoal)
    if err := c.BodyParser(goal); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    // Get user_id from request body
    userIDStr := c.FormValue("user_id")
    if userIDStr == "" {
        // Try to get from JSON body if not in form data
        var body map[string]interface{}
        if err := c.BodyParser(&body); err == nil {
            if id, ok := body["user_id"].(string); ok {
                userIDStr = id
            }
        }
    }

    // Convert user_id string to ObjectID
    userID, err := primitive.ObjectIDFromHex(userIDStr)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid user ID format",
            "details": err.Error(),
        })
    }

    goal.UserID = userID
    goal.ID = primitive.NewObjectID()
    goal.CreatedAt = time.Now()
    goal.UpdatedAt = time.Now()

    collection := database.GetCollection("health_goals")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err = collection.InsertOne(ctx, goal)
    if err != nil {
        log.Printf("Error saving goal: %v", err)
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to create health goal",
        })
    }

    return c.Status(201).JSON(goal)
}

// GetHealthGoal retrieves a health goal by user ID
func GetHealthGoal(c *fiber.Ctx) error {
    userID := c.Params("user_id")
    objectID, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid user ID format",
        })
    }

    collection := database.GetCollection("health_goals")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var goal models.HealthGoal
    err = collection.FindOne(ctx, bson.M{"user_id": objectID}).Decode(&goal)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{
            "error": "Health goal not found",
        })
    }

    return c.JSON(goal)
}

// UpdateHealthGoal updates an existing health goal
func UpdateHealthGoal(c *fiber.Ctx) error {
    userID := c.Params("user_id")
    objectID, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid user ID format",
        })
    }

    goal := new(models.HealthGoal)
    if err := c.BodyParser(goal); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    goal.UpdatedAt = time.Now()

    collection := database.GetCollection("health_goals")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    result, err := collection.UpdateOne(
        ctx,
        bson.M{"user_id": objectID},
        bson.M{"$set": goal},
    )
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to update health goal",
        })
    }

    if result.MatchedCount == 0 {
        return c.Status(404).JSON(fiber.Map{
            "error": "Health goal not found",
        })
    }

    return c.JSON(fiber.Map{
        "message": "Health goal updated successfully",
    })
}

func GetHealthGoalsByUserId(c *fiber.Ctx) error {
    userID, err := primitive.ObjectIDFromHex(c.Params("userId"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid user ID format",
        })
    }

    collection := database.GetCollection("health_goals")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Find the most recent health goal for the user
    var healthGoal models.HealthGoal
    err = collection.FindOne(ctx, bson.M{
        "user_id": userID,
    }).Decode(&healthGoal)

    if err != nil {
        if err == mongo.ErrNoDocuments {
            return c.Status(404).JSON(fiber.Map{
                "error": "No health goals found for this user",
            })
        }
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch health goals",
        })
    }

    return c.JSON(healthGoal)
}