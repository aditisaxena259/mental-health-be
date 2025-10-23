package routes

import (
	"github.com/aditisaxena259/mental-health-be/controllers"
	"github.com/aditisaxena259/mental-health-be/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// -------------------------------
	// PUBLIC ROUTES (no JWT required)
	// -------------------------------
	api.Post("/signup", controllers.Signup)
	api.Post("/login", controllers.Login)
	api.Post("/logout", controllers.Logout)

	// -------------------------------
	// PROTECTED ROUTES (JWT required)
	// -------------------------------
	protected := api.Group("/", middlewares.ProtectRoute)

	// --- Student routes ---
	student := protected.Group("/student", middlewares.RequireRole("student"))
	student.Post("/complaints", controllers.CreateComplaint)
	student.Get("/complaints", controllers.GetAllComplaints)
	student.Post("/apologies", controllers.SubmitApology)
	student.Get("/apologies", controllers.GetStudentApologies)

	// --- Admin routes ---
	admin := protected.Group("/admin", middlewares.RequireRole("admin"))
	admin.Put("/complaints/:id/status", controllers.UpdateComplaintStatus)
	admin.Delete("/complaints/:id", controllers.DeleteComplaint)
	admin.Put("/apologies/:id/review", controllers.ReviewApology)

	// --- Shared (any logged-in user) ---
	protected.Get("/metrics/status-summary", controllers.GetStatus)
	protected.Get("/metrics/resolution-rate", controllers.GetResolutionRate)
	protected.Get("/metrics/pending-count", controllers.GetPendingComplaint)

	// --- Complaint Timeline ---
	protected.Post("/complaints/:id/timeline", controllers.AddTimelineEntry)
	protected.Get("/complaints/:id/timeline", controllers.GetTimeline)
}
