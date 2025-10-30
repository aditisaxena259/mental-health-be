package main

import (
	"log"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/aditisaxena259/mental-health-be/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Warning: No .env file found")
	}

	// Connect to PostgreSQL
	if err := config.ConnectDatabase(); err != nil {
		log.Fatal("❌ Failed to connect to the database:", err)
	}
	log.Println("✅ Connected to PostgreSQL!")

	models.AutoMigrateAll()
	models.SeedData()
	log.Println("📦 Database migrations completed successfully!")

	// Initialize Fiber app
	app := fiber.New()

	// Enable CORS Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // frontend origin
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Setup API routes
	routes.SetupRoutes(app)

	// Start server
	log.Println("🚀 Server running at http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
