package main

import (
	"log"
	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/aditisaxena259/mental-health-be/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: No .env file found")
    }

    // Connect to the database (Using GORM)
    if err := config.ConnectDatabase(); err != nil {
        log.Fatal("Failed to connect to the database:", err)
    }

    // Run database migrations
    if err := config.DB.AutoMigrate(&models.User{}, &models.Complaint{}); err != nil {
        log.Fatal("Failed to migrate database:", err)
    }

    log.Println("Database migrated successfully!")

    // Initialize Fiber app
    app := fiber.New()
    routes.SetupRoutes(app)

    // Start the server
    log.Fatal(app.Listen(":8080"))
}
