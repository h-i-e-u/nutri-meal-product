package handlers

import (
	"context"
	"fmt"
	"log"
	"nitri-meal-backend/database"
	"nitri-meal-backend/models"
	"nitri-meal-backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateUser(c *fiber.Ctx) error {
    user := new(models.User)
    if err := c.BodyParser(user); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    collection := database.GetCollection("users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Check if user already exists by email
    var existingUser models.User
    err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
    if err == nil {
        // User exists, return the existing user
        return c.Status(200).JSON(fiber.Map{
            "id":      existingUser.ID.Hex(),
            "email":   existingUser.Email,
            "name":    existingUser.Name,
            "picture": existingUser.Picture,
        })
    }

    if err != mongo.ErrNoDocuments {
        // Handle other database errors
        return c.Status(500).JSON(fiber.Map{
            "error": "Database error",
        })
    }

    // User does not exist, create a new user
    user.ID = primitive.NewObjectID()
    user.CreatedAt = time.Now()
    user.UpdatedAt = time.Now()

    _, err = collection.InsertOne(ctx, user)
    if err != nil {
        // Handle duplicate key error (race condition)
        if mongo.IsDuplicateKeyError(err) {
            err = collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
            if err == nil {
                return c.Status(200).JSON(fiber.Map{
                    "id":      existingUser.ID.Hex(),
                    "email":   existingUser.Email,
                    "name":    existingUser.Name,
                    "picture": existingUser.Picture,
                })
            }
        }
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to create user",
        })
    }

    return c.Status(201).JSON(fiber.Map{
        "id":      user.ID.Hex(),
        "email":   user.Email,
        "name":    user.Name,
        "picture": user.Picture,
    })
}

func GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
    
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

func UpdateUser(c *fiber.Ctx) error {
    userId := c.Params("id")
    objectId, err := primitive.ObjectIDFromHex(userId)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    collection := database.GetCollection("users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Get existing user
    var existingUser models.User
    err = collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&existingUser)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }

    // Handle file upload
    file, err := c.FormFile("picture")
    if err == nil && file != nil {
        // Delete old image if exists
        if existingUser.DeleteHash != "" {
            err = utils.DeleteFromImgur(existingUser.DeleteHash)
            if err != nil {
                // Log the error but continue - don't block update if delete fails
                fmt.Printf("Failed to delete old image: %v\n", err)
            }
        }

        // Upload new image
        result, err := utils.UploadToImgur(file)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to upload image"})
        }

        // Update user with new image info
        update := bson.M{
            "$set": bson.M{
                "picture":    (*result)["link"].(string),
                "deleteHash": (*result)["deleteHash"].(string),
                "updatedAt": time.Now(),
            },
        }

        _, err = collection.UpdateOne(ctx, bson.M{"_id": objectId}, update)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to update user"})
        }
    }

    // Update other fields if provided
    var updateData map[string]interface{}
    if err := c.BodyParser(&updateData); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
    }

    // 
    delete(updateData, "picture") // Remove picture from regular updates
    if len(updateData) > 0 {
        updateData["updated_at"] = time.Now()
        update := bson.M{"$set": updateData}
        _, err = collection.UpdateOne(ctx, bson.M{"_id": objectId}, update)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to update user"})
        }
    }

    // Return the updated user data 
    var updatedUser models.User
    err = collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&updatedUser)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch updated user"})
    }

    return c.JSON(updatedUser)
}

func GetUserByEmail(c *fiber.Ctx) error {
    email := c.Query("email")
    if email == "" {
        return c.Status(400).JSON(fiber.Map{
            "error": "Email is required",
        })
    }

    collection := database.GetCollection("users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var user models.User
    err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return c.Status(404).JSON(fiber.Map{
                "error": "User not found",
            })
        }
        return c.Status(500).JSON(fiber.Map{
            "error": "Database error",
        })
    }

    return c.Status(200).JSON(fiber.Map{
        "_id":     user.ID,
        "email":   user.Email,
        "name":    user.Name,
        "picture": user.Picture,
    })
}

func UpdateUserPicture(c *fiber.Ctx) error {
    userId := c.Params("id")
    objectId, err := primitive.ObjectIDFromHex(userId)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
    }

    // Get file from request
    file, err := c.FormFile("picture")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
    }

    // Get existing user for delete hash
    collection := database.GetCollection("users")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var existingUser models.User
    err = collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&existingUser)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }

    // Delete old image if exists
    if existingUser.DeleteHash != "" {
        err = utils.DeleteFromImgur(existingUser.DeleteHash)
        if err != nil {
            // Log error but continue
            log.Printf("[ERROR] Failed to delete old image: %v\n", err)
        }
    }

    // Upload new image
    result, err := utils.UploadToImgur(file)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to upload image"})
    }

    // Update user with new image info
    update := bson.M{
        "$set": bson.M{
            "picture":    (*result)["link"].(string),
            "deleteHash": (*result)["deleteHash"].(string),
            "updatedAt":  time.Now(),
        },
    }

    _, err = collection.UpdateOne(ctx, bson.M{"_id": objectId}, update)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to update user"})
    }

    // Return updated user
    var updatedUser models.User
    err = collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&updatedUser)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch updated user"})
    }

    return c.JSON(updatedUser)
}
