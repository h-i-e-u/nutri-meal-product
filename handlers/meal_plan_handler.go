package handlers

import (
	"context"
	"math/rand"
	"nitri-meal-backend/database"
	"nitri-meal-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetMealPlansByUserID retrieves meal plans for a specific user
func GetMealPlansByUserID(c *fiber.Ctx) error {
    userID := c.Params("userId")
    if userID == "" {
        return c.Status(400).JSON(fiber.Map{
            "error": "User ID is required",
        })
    }

    collection := database.GetCollection("meal_plans")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var mealPlans []models.MealPlan
    cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch meal plans",
        })
    }
    defer cursor.Close(ctx)

    if err := cursor.All(ctx, &mealPlans); err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to decode meal plans",
        })
    }

    return c.JSON(mealPlans)
}

// CreateMealPlan creates a new meal plan with random recipes
func CreateMealPlan(c *fiber.Ctx) error {
    // Parse request body
    mealPlan := new(models.MealPlan)
    if err := c.BodyParser(mealPlan); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    collection := database.GetCollection("meal_plans")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Check if meal plan already exists for this date and user
    existingPlan := new(models.MealPlan)
    err := collection.FindOne(ctx, bson.M{
        "user_id": mealPlan.UserID,
        "date": mealPlan.Date,
    }).Decode(existingPlan)

    if err == nil {
        // Plan already exists for this date
        return c.Status(400).JSON(fiber.Map{
            "error": "A meal plan already exists for this date",
        })
    } else if err != mongo.ErrNoDocuments {
        // Some other error occurred
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to check existing meal plans",
        })
    }

    // Get recipes collection
    recipesCollection := database.GetCollection("recipes")

    // Get all recipes
    var recipes []models.Recipe
    cursor, err := recipesCollection.Find(ctx, bson.M{})
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch recipes",
        })
    }
    defer cursor.Close(ctx)

    if err := cursor.All(ctx, &recipes); err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to decode recipes",
        })
    }

    // Randomly select 3 recipes
    rand.Seed(time.Now().UnixNano())
    rand.Shuffle(len(recipes), func(i, j int) {
        recipes[i], recipes[j] = recipes[j], recipes[i]
    })

    selectedRecipes := recipes[:3]
    recipeIDs := make([]int, 3)

    // Create meal plan with random recipes
    mealPlan.ID = primitive.NewObjectID()
    mealPlan.Meal = models.Meals{
        Breakfast: selectedRecipes[0].Name,
        Lunch:     selectedRecipes[1].Name,
        Dinner:    selectedRecipes[2].Name,
    }

    // Store recipe IDs
    for i, recipe := range selectedRecipes {
        recipeIDs[i] = recipe.RecipeID
    }
    mealPlan.Recipes = recipeIDs

    // Get the next plan ID
    planID, err := getNextPlanID(ctx)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to generate plan ID",
        })
    }
    mealPlan.PlanID = planID

    // Save to database
    _, err = collection.InsertOne(ctx, mealPlan)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to create meal plan",
        })
    }

    return c.Status(201).JSON(mealPlan)
}

// Helper function to get next plan ID
func getNextPlanID(ctx context.Context) (int, error) {
    collection := database.GetCollection("meal_plans")
    
    // Find the meal plan with highest plan_id
    opts := options.FindOne().SetSort(bson.M{"id": -1})
    var lastPlan models.MealPlan
    
    err := collection.FindOne(ctx, bson.M{}, opts).Decode(&lastPlan)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return 1, nil // Start with ID 1 if no plans exist
        }
        return 0, err
    }
    
    return lastPlan.PlanID + 1, nil
}