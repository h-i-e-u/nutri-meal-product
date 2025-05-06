package handlers

import (
	"context"
	"nitri-meal-backend/database"
	"nitri-meal-backend/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetAllRecipes retrieves all recipes with pagination
func GetAllRecipes(c *fiber.Ctx) error {
    collection := database.GetCollection("recipes")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Parse pagination parameters
    page := c.QueryInt("page", 1)
    limit := c.QueryInt("limit", 10)
    skip := (page - 1) * limit

    // Setup options for pagination
    findOptions := options.Find().
        SetSkip(int64(skip)).
        SetLimit(int64(limit))

    // Query recipes
    cursor, err := collection.Find(ctx, bson.M{}, findOptions)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch recipes",
        })
    }
    defer cursor.Close(ctx)

    var recipes []models.Recipe
    if err := cursor.All(ctx, &recipes); err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to decode recipes",
        })
    }

    // Get total count for pagination
    total, err := collection.CountDocuments(ctx, bson.M{})
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to count recipes",
        })
    }

    return c.JSON(fiber.Map{
        "recipes": recipes,
        "total": total,
        "page": page,
        "limit": limit,
    })
}

// GetRecipeByID handles both MongoDB ObjectID and numeric ID
func GetRecipeByID(c *fiber.Ctx) error {
    idParam := c.Params("id")
    
    collection := database.GetCollection("recipes")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var recipe models.Recipe
    var err error

    // Try to parse as MongoDB ObjectID first
    if objectID, err := primitive.ObjectIDFromHex(idParam); err == nil {
        err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&recipe)
    } else {
        // If not ObjectID, try as numeric ID
        if numID, err := strconv.Atoi(idParam); err == nil {
            err = collection.FindOne(ctx, bson.M{"id": numID}).Decode(&recipe)
        } else {
            return c.Status(400).JSON(fiber.Map{
                "error": "Invalid recipe ID format",
            })
        }
    }

    if err != nil {
        return c.Status(404).JSON(fiber.Map{
            "error": "Recipe not found",
        })
    }

    return c.JSON(recipe)
}

// GetRecipesByCategory retrieves recipes by category
func GetRecipesByCategory(c *fiber.Ctx) error {
    category := c.Query("category")
    if category == "" {
        return c.Status(400).JSON(fiber.Map{
            "error": "Category is required",
        })
    }

    collection := database.GetCollection("recipes")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cursor, err := collection.Find(ctx, bson.M{"category": category})
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to fetch recipes",
        })
    }
    defer cursor.Close(ctx)

    var recipes []models.Recipe
    if err := cursor.All(ctx, &recipes); err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to decode recipes",
        })
    }

    return c.JSON(recipes)
}

// getNextRecipeID retrieves the next recipe ID
func getNextRecipeID(ctx context.Context, collection *mongo.Collection) (int, error) {
    // Find the recipe with highest recipe_id
    opts := options.FindOne().SetSort(bson.M{"recipe_id": -1})
    var lastRecipe models.Recipe
    
    err := collection.FindOne(ctx, bson.M{}, opts).Decode(&lastRecipe)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return 1, nil // Start with ID 1 if no recipes exist
        }
        return 0, err
    }
    
    return lastRecipe.RecipeID + 1, nil
}

// CreateRecipe creates a new recipe with auto-incrementing recipe_id
func CreateRecipe(c *fiber.Ctx) error {
    collection := database.GetCollection("recipes")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var recipe models.Recipe
    if err := c.BodyParser(&recipe); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid recipe data",
        })
    }

    // Get next recipe ID
    nextID, err := getNextRecipeID(ctx, collection)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to generate recipe ID",
        })
    }

    // Set the recipe ID
    recipe.RecipeID = nextID
    recipe.ID = primitive.NewObjectID()

    if _, err := collection.InsertOne(ctx, recipe); err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to create recipe",
        })
    }

    return c.Status(201).JSON(recipe)
}