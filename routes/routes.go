package routes

import (
	"github.com/aditisaxena259/mental-health-be/controllers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/signup", controllers.Signup)
	api.Post("/login", controllers.Login)
	api.Post("/logout", controllers.Logout)

	api.Post("/complaints", controllers.CreateComplaint)
	api.Get("/complaints", controllers.GetAllComplaints)
	api.Get("/complaints/student", controllers.GetStudentComplaints)
	api.Put("/complaints/:id/status", controllers.UpdateComplaintStatus)
	api.Delete("/complaints/:id", controllers.DeleteComplaint)

	api.Post("/apologies", controllers.SubmitApology)
	api.Get("/apologies", controllers.GetAllApologies)
	api.Get("/apologies/student", controllers.GetStudentApologies)
	api.Put("/apologies/:id/status", controllers.UpdateApologyStatus)
	api.Delete("/apologies/:id", controllers.DeleteApology)
}
