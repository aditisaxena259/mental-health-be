package controllers

import (
	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
)

// STUDENT - Submit Apology Letter
func SubmitApology(c *fiber.Ctx) error {
	var apology models.ApologyLetter

	if err := c.BodyParser(&apology); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	apology.Status = "submitted"
	config.DB.Create(&apology)

	return c.JSON(fiber.Map{"message": "Apology letter submitted successfully"})
}

// STUDENT - Get Own Apologies
func GetStudentApologies(c *fiber.Ctx) error {
	studentID := c.Query("student_id")
	if studentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Student ID required"})
	}

	var apologies []models.ApologyLetter
	config.DB.Where("student_id = ?", studentID).Find(&apologies)

	return c.JSON(apologies)
}

// ADMIN - Get All Apologies
func GetAllApologies(c *fiber.Ctx) error {
	var apologies []models.ApologyLetter
	config.DB.Find(&apologies)

	return c.JSON(apologies)
}

// ADMIN - Update Apology Status
func UpdateApologyStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var updateData struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	var apology models.ApologyLetter
	if err := config.DB.First(&apology, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Apology not found"})
	}

	apology.Status = models.ApologyStatus(updateData.Status)
	config.DB.Save(&apology)

	return c.JSON(fiber.Map{"message": "Apology status updated"})
}

// ADMIN - Delete Apology
func DeleteApology(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := config.DB.Delete(&models.ApologyLetter{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete apology"})
	}

	return c.JSON(fiber.Map{"message": "Apology letter deleted successfully"})
}
