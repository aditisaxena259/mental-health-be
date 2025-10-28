package controllers

import (
	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// 🧑‍🎓 STUDENT — Submit Apology Letter
func SubmitApology(c *fiber.Ctx) error {
	// Extract logged-in student ID from JWT
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: missing user ID"})
	}

	var input struct {
		Type        models.ApologyType `json:"type"`
		Message     string             `json:"message"`
		Description string             `json:"description"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input", "details": err.Error()})
	}

	if input.Message == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Message field is required"})
	}

	// ✅ Convert userID string → uuid.UUID
	studentUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	apology := models.Apology{
		StudentID:   studentUUID,
		ApologyType: input.Type,
		Message:     input.Message,
		Description: input.Description,
		Status:      models.ApologySubmitted,
	}

	if err := config.DB.Create(&apology).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to submit apology", "details": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Apology letter submitted successfully",
		"data":    apology,
	})
}

// 🧑‍🎓 STUDENT — Get Own Apologies
func GetStudentApologies(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var apologies []models.Apology
	if err := config.DB.Where("student_id = ?", userID).Order("created_at desc").Find(&apologies).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch apologies"})
	}

	if len(apologies) == 0 {
		return c.JSON(fiber.Map{"message": "No apologies found", "data": []models.Apology{}})
	}

	return c.JSON(fiber.Map{"count": len(apologies), "data": apologies})
}

// 🧑‍💼 ADMIN — Get All or Filtered Apologies
func GetApologies(c *fiber.Ctx) error {
	var apologies []models.Apology
	query := config.DB

	if apologyType := c.Query("type"); apologyType != "" {
		query = query.Where("apology_type = ?", apologyType)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at desc").Find(&apologies).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch apologies"})
	}

	return c.JSON(fiber.Map{"count": len(apologies), "data": apologies})
}

// 🧑‍💼 ADMIN — Get Apology by ID
func GetApologyByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var apology models.Apology

	if err := config.DB.First(&apology, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Apology not found"})
	}

	return c.JSON(apology)
}

// 🧑‍💼 ADMIN — Review or Update Apology Status (Transactional)
func ReviewApology(c *fiber.Ctx) error {
	id := c.Params("id")

	var input struct {
		Status  models.ApologyStatus `json:"status"`
		Comment string               `json:"comment"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var apology models.Apology
	if err := tx.First(&apology, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "Apology not found"})
	}

	apology.Status = input.Status
	apology.Comment = input.Comment

	if err := tx.Save(&apology).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update apology"})
	}

	tx.Commit()
	return c.JSON(fiber.Map{
		"message": "Apology reviewed successfully",
		"data":    apology,
	})
}

// 🧾 ADMIN — Pending Count
func GetPendingApology(c *fiber.Ctx) error {
	var count int64
	if err := config.DB.Model(&models.Apology{}).Where("status = ?", models.ApologySubmitted).Count(&count).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count pending apologies"})
	}
	return c.JSON(fiber.Map{"pending_count": count})
}
