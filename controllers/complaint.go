package controllers

import (
	"fmt"
	"time"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateComplaint(c *fiber.Ctx) error {
	var complaint models.Complaint

	if err := c.BodyParser(&complaint); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// ✅ Extract user ID from JWT (automatically set by middleware)
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: missing user ID"})
	}

	complaint.UserID = uuid.MustParse(userID)
	complaint.Status = "open" // Default value

	// Save to DB
	if err := config.DB.Create(&complaint).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create complaint",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "Complaint submitted successfully"})
}




// ADMIN - Get All Complaints (Optional Filter by Type)
// func GetAllComplaints(c *fiber.Ctx) error {
// 	var complaints []models.Complaint
// 	complaintType := c.Query("type")

// 	query := config.DB.Model(&models.Complaint{})
// 	if complaintType != "" {
// 		query = query.Where("type = ?", complaintType)
// 	}

// 	query.Find(&complaints)

// 	return c.JSON(complaints)
// }



// GetAllComplaints returns all complaints for the logged-in user.
// Admins can view all complaints, students see only their own.
func GetAllComplaints(c *fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("user_id").(string)

	var complaints []models.Complaint

	query := config.DB.
		Preload("User").
		Preload("Student").
		Preload("Timeline")

	// Try preloading Attachments — if it fails (missing table), skip it.
	err := query.Preload("Attachments").Order("created_at DESC").Find(&complaints).Error
	if err != nil && err.Error() == `ERROR: relation "attachments" does not exist (SQLSTATE 42P01)` {
		// Retry without attachments
		err = config.DB.
			Preload("User").
			Preload("Student").
			Preload("Timeline").
			Order("created_at DESC").
			Find(&complaints).Error
	}

	// Apply role filter if needed
	if role == "student" && userID != "" {
		filtered := []models.Complaint{}
		for _, c := range complaints {
			if c.UserID.String() == userID {
				filtered = append(filtered, c)
			}
		}
		complaints = filtered
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to fetch complaints",
			"details": err.Error(),
		})
	}

	if len(complaints) == 0 {
		return c.JSON(fiber.Map{
			"message": "No complaints found",
			"data":    []models.Complaint{},
		})
	}

	return c.JSON(fiber.Map{
		"count": len(complaints),
		"data":  complaints,
	})
}


// ADMIN - Update Complaint Status
func UpdateComplaintStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var input struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	var complaint models.Complaint
	if err := config.DB.First(&complaint, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Complaint not found"})
	}

	tx := config.DB.Begin()
	complaint.Status = models.ComplaintStatus(input.Status)
	if err := tx.Save(&complaint).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update status"})
	}

	timeline := models.TimelineEntry{
		ComplaintID: complaint.ID,
		Author:      "Warden/Admin",
		Message:     fmt.Sprintf("Status changed to %s", input.Status),
		Timestamp:   time.Now(),
	}
	tx.Create(&timeline)

	tx.Commit()
	return c.JSON(fiber.Map{"message": "Status updated"})
}


// ADMIN - Delete Complaint
func DeleteComplaint(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := config.DB.Delete(&models.Complaint{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete complaint"})
	}

	return c.JSON(fiber.Map{"message": "Complaint deleted successfully"})
}
// func GetPendingComplaint(c *fiber.Ctx)error{
//     var count int64;
//     err:= config.DB.Model(&models.Complaint{}).
//     Where("status = ?", "pending").
//     Count(&count).Error
//     if err != nil {
//         return c.Status(500).JSON(fiber.Map{
//             "error": "Failed to count pending complaints",
//         })
//     }
//     return c.JSON(fiber.Map{
//         "pending_count": count,
//     })
// }
// func GetResolutionRate(c *fiber.Ctx)error{
// 	var resolvedCount int64
// 	var totalCount int64

// 	err := config.DB.Model(&models.Complaint{}).
// 		Where("status = ?", "resolved").
// 		Count(&resolvedCount).Error
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Failed to count resolved complaints",
// 		})
// 	}

// 	err = config.DB.Model(&models.Complaint{}).Count(&totalCount).Error
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"error": "Failed to count total complaints",
// 		})
// 	}
// 	if totalCount==0{
// 		return c.JSON(fiber.Map{
// 			"resolution_rate": 0.0,
// 		})
// 	}
// 	rate:=float64(resolvedCount) / float64(totalCount) * 100
// 	return c.JSON(fiber.Map{
// 		"resolution_rate": rate,
// 	})
// }

// func GetStatus(c *fiber.Ctx) error {
// 	var resolved, open, inProgress int64

// 	config.DB.Model(&models.Complaint{}).Where("status = ?", "resolved").Count(&resolved)
// 	config.DB.Model(&models.Complaint{}).Where("status = ?", "open").Count(&open)
// 	config.DB.Model(&models.Complaint{}).Where("status = ?", "inprogress").Count(&inProgress)

// 	return c.JSON(fiber.Map{
// 		"resolved":   resolved,
// 		"open":       open,
// 		"inprogress": inProgress,
// 	})
// }

func GetComplaintsByType(c *fiber.Ctx) error {
	var complaints []models.Complaint
	complaintType := c.Query("type")
	query := config.DB.Model(&models.Complaint{})
	if complaintType != "" {
		query = query.Where("type = ?", complaintType)
	}
	if err := query.Order("created_at desc").Find(&complaints).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch complaints",
		})
	}

	return c.JSON(complaints)
}

func GetComplaintbyID(c *fiber.Ctx) error {
	complaintID := c.Params("id")
	var complaint models.Complaint

	// First try with Attachments
	err := config.DB.Preload("User").
		Preload("Timeline").
		Preload("Attachments").
		First(&complaint, "id = ?", complaintID).Error

	// If attachments table missing, try again without it
	if err != nil && err.Error() == `ERROR: relation "attachments" does not exist (SQLSTATE 42P01)` {
		err = config.DB.Preload("User").
			Preload("Timeline").
			First(&complaint, "id = ?", complaintID).Error
	}

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error":   "Complaint not found",
			"details": err.Error(),
		})
	}

	// Fetch user's past complaints
	var pastComplaints []models.Complaint
	config.DB.Where("user_id = ? AND id != ?", complaint.UserID, complaint.ID).
		Find(&pastComplaints)

	response := fiber.Map{
		"id":          complaint.ID,
		"title":       complaint.Title,
		"type":        complaint.Type,
		"description": complaint.Description,
		"status":      complaint.Status,
		"created_at":  complaint.CreatedAt,
		"user": fiber.Map{
			"id":      complaint.User.ID,
			"name":    complaint.User.Name,
			"email":   complaint.User.Email,
			"hostel":  complaint.Student.Hostel,
			"room_no": complaint.Student.RoomNo,
		},
		"attachments": complaint.Attachments, // if table exists, filled; else empty
		"timeline":    complaint.Timeline,
		"past_complaints": pastComplaints,
	}

	return c.JSON(response)
}
