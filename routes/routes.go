package routes

import (
	"github.com/aditisaxena259/mental-health-be/controllers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
    api := app.Group("/api")

    // Auth Routes
    api.Post("/signup", controllers.Signup)
    api.Post("/login", controllers.Login)

    // Complaint Routes
    api.Post("/complaints", controllers.CreateComplaint)
    api.Get("/complaints", controllers.GetComplaints)
}
