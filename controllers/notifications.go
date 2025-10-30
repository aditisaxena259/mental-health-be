package controllers

import (
	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GET /api/admin/notifications - list notifications for the logged-in admin
func GetNotifications(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	uid := uuid.MustParse(userID)
	var notes []models.Notification
	if err := config.DB.Where("admin_id = ?", uid).Order("created_at desc").Find(&notes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch notifications"})
	}
	return c.JSON(fiber.Map{"count": len(notes), "data": notes})
}

// POST /api/admin/notifications/:id/read - mark as read
func MarkNotificationRead(c *fiber.Ctx) error {
	id := c.Params("id")
	var n models.Notification
	if err := config.DB.First(&n, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "notification not found"})
	}
	n.IsRead = true
	config.DB.Save(&n)
	return c.JSON(fiber.Map{"message": "marked read"})
}
