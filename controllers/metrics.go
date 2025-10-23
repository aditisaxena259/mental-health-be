package controllers

import (
	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
)

// GET /metrics/status-summary
func GetStatus(c *fiber.Ctx) error {
	var open, inprogress, resolved int64

	config.DB.Model(&models.Complaint{}).Where("status = ?", "open").Count(&open)
	config.DB.Model(&models.Complaint{}).Where("status = ?", "inprogress").Count(&inprogress)
	config.DB.Model(&models.Complaint{}).Where("status = ?", "resolved").Count(&resolved)

	return c.JSON(fiber.Map{
		"open":        open,
		"inprogress":  inprogress,
		"resolved":    resolved,
		"total":       open + inprogress + resolved,
	})
}

// GET /metrics/resolution-rate
func GetResolutionRate(c *fiber.Ctx) error {
	var resolved, total int64
	config.DB.Model(&models.Complaint{}).Where("status = ?", "resolved").Count(&resolved)
	config.DB.Model(&models.Complaint{}).Count(&total)

	if total == 0 {
		return c.JSON(fiber.Map{"resolution_rate": 0})
	}
	rate := (float64(resolved) / float64(total)) * 100
	return c.JSON(fiber.Map{"resolution_rate": rate})
}

// GET /metrics/pending-count
func GetPendingComplaint(c *fiber.Ctx) error {
	var count int64
	config.DB.Model(&models.Complaint{}).
		Where("status = ?", "open").
		Count(&count)
	return c.JSON(fiber.Map{"pending_count": count})
}
