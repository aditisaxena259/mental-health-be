package controllers

import (
	"time"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// POST /complaints/:id/timeline
func AddTimelineEntry(c *fiber.Ctx) error {
	complaintID := c.Params("id")
	userID := c.Locals("user_id").(string)
	role := c.Locals("role").(string)

	var input struct {
		Message string `json:"message"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	entry := models.TimelineEntry{
		ID:          uuid.New(),
		ComplaintID: uuid.MustParse(complaintID),
		Author:      role + ":" + userID,
		Message:     input.Message,
		Timestamp:   time.Now(),
	}

	if err := config.DB.Create(&entry).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to add timeline entry"})
	}

	return c.JSON(entry)
}

// GET /complaints/:id/timeline
func GetTimeline(c *fiber.Ctx) error {
	id := c.Params("id")
	var timeline []models.TimelineEntry
	if err := config.DB.Where("complaint_id = ?", id).
		Order("timestamp asc").
		Find(&timeline).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load timeline"})
	}
	return c.JSON(timeline)
}
