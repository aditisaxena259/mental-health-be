package controllers

import (
	"os"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GET /api/notifications - list notifications for logged-in user (admin or student)
func GetNotifications(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	uid := uuid.MustParse(userID)
	var notes []models.Notification
	if err := config.DB.Where("user_id = ?", uid).Order("created_at desc").Find(&notes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch notifications"})
	}
	unread := 0
	for _, n := range notes {
		if !n.IsRead {
			unread++
		}
	}
	return c.JSON(fiber.Map{"unreadCount": unread, "data": notes})
}

// PATCH /api/notifications/:id/read - mark a notification as read for the logged-in user
func MarkNotificationRead(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	uid := uuid.MustParse(userID)

	var n models.Notification
	if err := config.DB.First(&n, "id = ? AND user_id = ?", id, uid).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "notification not found"})
	}
	n.IsRead = true
	config.DB.Save(&n)
	return c.JSON(fiber.Map{"message": "marked read"})
}

// PATCH /api/notifications/read-all - mark all notifications as read for the logged-in user
func MarkAllNotificationsRead(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	uid := uuid.MustParse(userID)
	config.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = false", uid).Updates(map[string]interface{}{"is_read": true})
	return c.JSON(fiber.Map{"message": "all marked read"})
}

// DELETE /api/notifications/:id - delete a notification for the logged-in user
func DeleteNotification(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	uid := uuid.MustParse(userID)
	if err := config.DB.Delete(&models.Notification{}, "id = ? AND user_id = ?", id, uid).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete"})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// DEV helper: return all notifications (only when DEV_MODE=true). Mounted under admin routes for quick debugging.
func DebugAllNotifications(c *fiber.Ctx) error {
	if os.Getenv("DEV_MODE") != "true" {
		return c.Status(403).JSON(fiber.Map{"error": "disabled"})
	}
	var notes []models.Notification
	if err := config.DB.Order("created_at desc").Find(&notes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch notifications", "details": err.Error()})
	}
	return c.JSON(fiber.Map{"count": len(notes), "data": notes})
}
