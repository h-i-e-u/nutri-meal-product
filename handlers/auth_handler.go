package handlers

import (
	"context"
	"log"
	"nitri-meal-backend/config"
	"nitri-meal-backend/database"
	"nitri-meal-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SignInRequest represents the sign-in request body
type SignInRequest struct {
	Email string `json:"email"`
}

// SignUpRequest represents the sign-up request body
type SignUpRequest struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// SignIn handles user sign-in
func SignIn(c *fiber.Ctx) error {
	var req SignInRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
			// "detail err": err.Error(),
		})
	}

	// Set session cookie
	store := config.GetStore()
	sess, err := store.Get(c)
	if err != nil {
		log.Printf("Session error: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Session error",
		})
	}

	sess.Set("user_id", user.ID.Hex())
	if err := sess.Save(); err != nil {
		log.Printf("Session save error: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Session save error",
		})
	}

	return c.JSON(fiber.Map{
		"_id":     user.ID,
		"email":   user.Email,
		"name":    user.Name,
		"picture": user.Picture,
	})
}

// SignUp handles user registration
func SignUp(c *fiber.Ctx) error {
	var req SignUpRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	collection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Email already exists",
		})
	}

	// Create new user
	newUser := models.User{
		ID:        primitive.NewObjectID(),
		Email:     req.Email,
		Name:      req.Name,
		Picture:   req.Picture,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = collection.InsertOne(ctx, newUser)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Set session cookie
	store := config.GetStore()
	sess, err := store.Get(c)
	if err != nil {
		log.Printf("Session error: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Session error",
		})
	}

	sess.Set("user_id", newUser.ID.Hex())
	if err := sess.Save(); err != nil {
		log.Printf("Session save error: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Session save error",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"_id":     newUser.ID,
		"email":   newUser.Email,
		"name":    newUser.Name,
		"picture": newUser.Picture,
	})
}

// SignOut handles user sign-out
func SignOut(c *fiber.Ctx) error {
	store := config.GetStore()
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Session error",
		})
	}

	if err := sess.Destroy(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to sign out",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Signed out successfully",
	})
}