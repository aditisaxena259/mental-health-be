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
	// Cloudinary health (keep public)
	api.Get("/health/cloudinary", controllers.CloudinaryPing)

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
	admin := protected.Group("/admin", middlewares.RequireRole("admin", "chief_admin"))

	// 🧾 Complaints
	admin.Get("/complaints", controllers.GetAllComplaintsAdmin)
	admin.Put("/complaints/:id/status", controllers.UpdateComplaintStatus)
	admin.Delete("/complaints/:id", controllers.DeleteComplaint)

	// ✉️ Apologies (admin/warden can see all student apologies)
	admin.Get("/apologies", controllers.GetApologies)              // View all or filter
	admin.Get("/apologies/:id", controllers.GetApologyByID)        // View specific apology
	admin.Put("/apologies/:id/review", controllers.ReviewApology)  // Review/accept/reject apology
	admin.Get("/apologies/pending", controllers.GetPendingApology) // Count pending apologies

	// 🔔 Notifications for admins
	admin.Get("/notifications", controllers.GetNotifications)
	admin.Post("/notifications/:id/read", controllers.MarkNotificationRead)
	// DEV: list all notifications for troubleshooting (only active when DEV_MODE=true)
	admin.Get("/notifications/debug", controllers.DebugAllNotifications)

	// Generic notification endpoints (for students and admins)
	protected.Get("/notifications", controllers.GetNotifications)
	protected.Patch("/notifications/:id/read", controllers.MarkNotificationRead)
	protected.Patch("/notifications/read-all", controllers.MarkAllNotificationsRead)
	protected.Delete("/notifications/:id", controllers.DeleteNotification)

	// -------------------------------
	// METRICS (Shared for logged-in users)
	// -------------------------------
	protected.Get("/metrics/status-summary", controllers.GetStatus)
	protected.Get("/metrics/resolution-rate", controllers.GetResolutionRate)
	protected.Get("/metrics/pending-count", controllers.GetPendingComplaint)

	// -------------------------------
	// Password reset (public)
	// -------------------------------
	api.Post("/forgot-password", controllers.ForgotPassword)
	api.Post("/reset-password", controllers.ResetPassword)

	// -------------------------------
	// COMPLAINT TIMELINE (Shared)
	// -------------------------------
	protected.Post("/complaints/:id/timeline", controllers.AddTimelineEntry)
	protected.Get("/complaints/:id/timeline", controllers.GetTimeline)

	// -------------------------------
	// Counseling / Slots
	// -------------------------------
	protected.Get("/counselors/:id/slots", controllers.ListCounselorSlots)
	// create slot (admin/counselor)
	protected.Post("/counselors/:id/slots", controllers.CreateCounselorSlot)
	// Admin books a slot for a student
	admin.Post("/counselors/:id/book", controllers.BookCounselorSlot)
	// Counselor updates session notes/progress (only counselor role)
	counselor := protected.Group("/counselor", middlewares.RequireRole("counselor"))
	counselor.Post("/sessions/:id/update", controllers.CounselorUpdateSession)

	// DEV helper to retrieve latest reset token (only when DEV_MODE=true)
	api.Get("/dev/reset-token", controllers.DevGetResetToken)

	// -------------------------------
	// Profile routes
	// -------------------------------
	student.Get("/profile", controllers.GetOwnProfile)
	admin.Get("/student/:student_identifier", controllers.GetStudentProfileAdmin)
}
