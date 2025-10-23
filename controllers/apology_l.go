package controllers

import (
	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
)

// STUDENT - Submit Apology Letter
func SubmitApology(c *fiber.Ctx) error {
	var apology models.Apology

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

	var apologies []models.Apology
	config.DB.Where("student_id = ?", studentID).Find(&apologies)

	return c.JSON(apologies)
}

// ADMIN - Get All Apologies
func GetApologiesbyID(c* fiber.Ctx) error{
    id:=c.Params("id");
    var apology models.Apology
    if err := config.DB.First(&apology, "id = ?", id).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{
            "error": "Apology not found",
        })
    }

    return c.JSON(apology)
}

func GetApologies(c *fiber.Ctx) error {
    var apologies []models.Apology
    query := config.DB

    if apologyType := c.Query("type"); apologyType != "" {
        query = query.Where("type = ?", apologyType)
    }

    if status := c.Query("status"); status != "" {
        query = query.Where("status = ?", status)
    }

    err := query.Order("created_at desc").Find(&apologies).Error
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch apologies"})
    }

    return c.JSON(apologies)
}

func ReviewApology(c *fiber.Ctx)error{
	id := c.Params("id")
    var apology models.Apology
	
	if err := config.DB.First(&apology, id).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Apology not found"})
    }
	type ReviewInput struct{
		Comment string `json:"comment"`
		Status models.ApologyStatus `json:"status"`
	}
	var input ReviewInput
    if err := c.BodyParser(&input); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
    }
	apology.Comment = input.Comment
    apology.Status = input.Status
    config.DB.Save(&apology)

    return c.JSON(apology)
}

func GetPendingApology(c *fiber.Ctx)error{
    var count int64;
    err:= config.DB.Model(&models.Apology{}).
    Where("status = ?", "pending").
    Count(&count).Error
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to count pending apologies",
        })
    }
    return c.JSON(fiber.Map{
        "pending_count": count,
    })
}

