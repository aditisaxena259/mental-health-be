package controllers

import (
	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
)

func CreateComplaint(c *fiber.Ctx) error {
    var complaint models.Complaint

    if err := c.BodyParser(&complaint); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
    }

    config.DB.Create(&complaint)

    return c.JSON(fiber.Map{"message": "Complaint registered successfully"})
}

func GetComplaints(c *fiber.Ctx) error {
    var complaints []models.Complaint
    complaintType := c.Query("type")

    query := config.DB.Model(&models.Complaint{})
    if complaintType != "" {
        query = query.Where("type = ?", complaintType)
    }

    query.Find(&complaints)

    return c.JSON(complaints)
}

func GetStudentComplaints(c *fiber.Ctx) error {
	studentID := c.Query("student_id")
	if studentID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Student ID required"})
	}

	var complaints []models.Complaint
	config.DB.Where("student_id = ?", studentID).Find(&complaints)

	return c.JSON(complaints)
}

// ADMIN - Get All Complaints (Optional Filter by Type)
func GetAllComplaints(c *fiber.Ctx) error {
	var complaints []models.Complaint
	complaintType := c.Query("type")

	query := config.DB.Model(&models.Complaint{})
	if complaintType != "" {
		query = query.Where("type = ?", complaintType)
	}

	query.Find(&complaints)

	return c.JSON(complaints)
}

// ADMIN - Update Complaint Status
func UpdateComplaintStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var updateData struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	var complaint models.Complaint
	if err := config.DB.First(&complaint, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Complaint not found"})
	}

	complaint.Status = models.StatusType(updateData.Status)
	config.DB.Save(&complaint)

	return c.JSON(fiber.Map{"message": "Complaint status updated"})
}

// ADMIN - Delete Complaint
func DeleteComplaint(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := config.DB.Delete(&models.Complaint{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete complaint"})
	}

	return c.JSON(fiber.Map{"message": "Complaint deleted successfully"})
}
