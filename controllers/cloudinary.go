package controllers

import (
	"github.com/aditisaxena259/mental-health-be/helpers"
	"github.com/gofiber/fiber/v2"
)

// CloudinaryPing checks if Cloudinary client can initialize with current env
func CloudinaryPing(c *fiber.Ctx) error {
	if _, err := helpers.InitCloudinary(); err != nil {
		return c.Status(200).JSON(fiber.Map{"ok": false, "error": err.Error()})
	}
	return c.Status(200).JSON(fiber.Map{"ok": true})
}
