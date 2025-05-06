package main

import (
	"log"
	"nitri-meal-backend/config"
	"nitri-meal-backend/database"
	"nitri-meal-backend/routes"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	database.Connect()
	defer database.Close()

	//  session store
	config.InitSession()

	// create  app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// middleware
	app.Use(logger.New())
	if os.Getenv("ENV") == "development" {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     "http://localhost:5173, https://localhost:5173", 
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
			AllowCredentials: true, 
			AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS", 
			MaxAge:           300,
		}))
	}

	// setup routes
	routes.SetupRoutes(app)

	// serve static files
	app.Static("/", "./client/dist")
	app.Get("*", func (c *fiber.Ctx) error {
		return c.SendFile("./client/dist/index.html")
	})


	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
