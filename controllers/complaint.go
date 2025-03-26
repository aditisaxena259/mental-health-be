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
