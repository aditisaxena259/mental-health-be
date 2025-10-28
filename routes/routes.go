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

	// -------------------------------
	// STUDENT ROUTES
	// -------------------------------
	student := protected.Group("/student", middlewares.RequireRole("student"))
	student.Post("/complaints", controllers.CreateComplaint)
	student.Get("/complaints", controllers.GetAllComplaints)

	// ✉️ Student Apologies
	student.Post("/apologies", controllers.SubmitApology)
	student.Get("/apologies", controllers.GetStudentApologies)

	// -------------------------------
	// ADMIN / WARDEN ROUTES
	// -------------------------------
	admin := protected.Group("/admin", middlewares.RequireRole("admin"))

	// 🧾 Complaints
	admin.Get("/complaints", controllers.GetAllComplaintsAdmin)
	admin.Put("/complaints/:id/status", controllers.UpdateComplaintStatus)
	admin.Delete("/complaints/:id", controllers.DeleteComplaint)

	// ✉️ Apologies (admin/warden can see all student apologies)
	admin.Get("/apologies", controllers.GetApologies)           // View all or filter
	admin.Get("/apologies/:id", controllers.GetApologyByID)     // View specific apology
	admin.Put("/apologies/:id/review", controllers.ReviewApology) // Review/accept/reject apology
	admin.Get("/apologies/pending", controllers.GetPendingApology) // Count pending apologies

	// -------------------------------
	// METRICS (Shared for logged-in users)
	// -------------------------------
	protected.Get("/metrics/status-summary", controllers.GetStatus)
	protected.Get("/metrics/resolution-rate", controllers.GetResolutionRate)
	protected.Get("/metrics/pending-count", controllers.GetPendingComplaint)

	// -------------------------------
	// COMPLAINT TIMELINE (Shared)
	// -------------------------------
	protected.Post("/complaints/:id/timeline", controllers.AddTimelineEntry)
	protected.Get("/complaints/:id/timeline", controllers.GetTimeline)
}
