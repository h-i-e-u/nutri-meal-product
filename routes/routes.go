package routes

import (
	"nitri-meal-backend/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// API group
	api := app.Group("/api")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/signin", handlers.SignIn)
	auth.Post("/signup", handlers.SignUp)
	auth.Post("/signout", handlers.SignOut)

	// User routes
	users := api.Group("/users")
	users.Get("/email", handlers.GetUserByEmail)
	users.Post("/", handlers.CreateUser)
	users.Get("/:id", handlers.GetUser)
	users.Put("/:id", handlers.UpdateUser)
	users.Put("/:id/picture", handlers.UpdateUserPicture) // Add this line

	// Health goal routes
	healthGoals := api.Group("/health-goals")
	healthGoals.Post("/", handlers.CreateHealthGoal)
	healthGoals.Get("/user/:userId", handlers.GetHealthGoalsByUserId)
	healthGoals.Get("/:user_id", handlers.GetHealthGoal)
	healthGoals.Put("/:user_id", handlers.UpdateHealthGoal)

	// reciepe routes
	recipes := api.Group("/recipes")
	recipes.Get("/", handlers.GetAllRecipes)
	recipes.Get("/:id", handlers.GetRecipeByID)
	recipes.Get("/category", handlers.GetRecipesByCategory)
	// recipes.Post("/", handlers.CreateRecipe) // later for nutritionist

	// Meal plan routes
	mealPlans := api.Group("/meal-plans")
	mealPlans.Get("/user/:userId", handlers.GetMealPlansByUserID)
	mealPlans.Post("/", handlers.CreateMealPlan)

	// Food log routes
	foodLogs := api.Group("/food-logs")
	foodLogs.Get("/user/:userId", handlers.GetFoodLogsByUserID)
	foodLogs.Post("/", handlers.CreateFoodLog)

	// Community routes
	community := api.Group("/community")
	community.Get("/posts", handlers.GetCommunityPosts)
	community.Post("/posts", handlers.CreateCommunityPost)
	community.Post("/posts/:postId/like", handlers.LikePost)  
	community.Delete("posts/:postId", handlers.DeleteCommunityPost)
}
