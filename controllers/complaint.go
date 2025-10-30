package controllers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 🧑‍🎓 STUDENT — Create Complaint
func CreateComplaint(c *fiber.Ctx) error {
	// Expect multipart/form-data: fields for title, type, description; files[] for attachments
	title := c.FormValue("title")
	ctype := c.FormValue("type")
	description := c.FormValue("description")
	priorityStr := c.FormValue("priority")

	if title == "" || ctype == "" || description == "" {
		return c.Status(400).JSON(fiber.Map{"error": "title, type and description are required"})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: missing user ID"})
	}

	complaint := models.Complaint{
		ID:          uuid.New(),
		Title:       title,
		Type:        models.ComplaintType(ctype),
		Description: description,
		UserID:      uuid.MustParse(userID),
		Status:      models.Open,
	}

	// set priority if provided
	if priorityStr != "" {
		complaint.Priority = models.ComplaintPriority(priorityStr)
	}

	// set student identifier from student's StudentModel (external id)
	var sm models.StudentModel
	if err := config.DB.Where("user_id = ?", complaint.UserID).First(&sm).Error; err == nil {
		complaint.StudentIdentifier = sm.StudentIdentifier
	}

	// Start transaction
	tx := config.DB.Begin()
	if err := tx.Create(&complaint).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create complaint", "details": err.Error()})
	}

	// NOTE: defer creating notifications until after the DB transaction commits
	// to avoid creating notifications for complaints that later rollback (e.g. attachment failures).

	// Handle attachments (optional). Field name: "attachments" (multiple)
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		files := form.File["attachments"]
		for _, fh := range files {
			if fh == nil {
				continue
			}
			if !isJPEG(fh) {
				tx.Rollback()
				return c.Status(400).JSON(fiber.Map{"error": "Only JPEG attachments are allowed"})
			}
			saved, saveErr := saveAttachmentFile(fh, complaint.ID)
			if saveErr != nil {
				tx.Rollback()
				return c.Status(500).JSON(fiber.Map{"error": "Failed to save attachment", "details": saveErr.Error()})
			}
			att := models.Attachment{
				ID:          uuid.New(),
				ComplaintID: complaint.ID,
				FileName:    fh.Filename,
				FileURL:     saved.PublicURL,
				Size:        fmt.Sprintf("%d", saved.Size),
				FilePath:    saved.Path,
			}
			if err := tx.Create(&att).Error; err != nil {
				tx.Rollback()
				return c.Status(500).JSON(fiber.Map{"error": "Failed to create attachment record"})
			}
		}
	}

	tx.Commit()
	// spawn notification creator after commit
	go func(comp models.Complaint) {
		var sm models.StudentModel
		if err := config.DB.Where("user_id = ?", comp.UserID).First(&sm).Error; err != nil {
			return
		}
		hostel := sm.Hostel
		var admins []models.User
		if hostel != "" {
			config.DB.Where("role = ? AND block = ?", models.Admin, hostel).Find(&admins)
		}
		var chiefs []models.User
		config.DB.Where("role = ?", models.ChiefAdmin).Find(&chiefs)
		// deduplicate by ID
		m := map[string]models.User{}
		for _, a := range admins {
			m[a.ID.String()] = a
		}
		for _, c := range chiefs {
			m[c.ID.String()] = c
		}
		admins = []models.User{}
		for _, v := range m {
			admins = append(admins, v)
		}

		for _, a := range admins {
			n := models.Notification{
				ID:      uuid.New(),
				AdminID: a.ID,
				Title:   "New Complaint Submitted",
				Body:    "A student has submitted a complaint: " + comp.Title,
				Link:    "/admin/complaints",
			}
			config.DB.Create(&n)
		}
	}(complaint)

	return c.JSON(fiber.Map{"message": "Complaint submitted successfully", "id": complaint.ID})
}

type savedFileInfo struct {
	Path      string
	PublicURL string
	Size      int64
}

func saveAttachmentFile(fh *multipart.FileHeader, complaintID uuid.UUID) (*savedFileInfo, error) {
	// ensure directory
	dir := "./uploads/attachments/" + complaintID.String()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	src, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dstPath := filepath.Join(dir, fh.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return nil, err
	}

	return &savedFileInfo{Path: dstPath, PublicURL: dstPath, Size: written}, nil
}

func isJPEG(fh *multipart.FileHeader) bool {
	f, err := fh.Open()
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	if n == 0 {
		return false
	}
	ct := http.DetectContentType(buf[:n])
	return ct == "image/jpeg" || ct == "image/jpg"
}

// 🧾 STUDENT + ADMIN — Get All Complaints
func GetAllComplaints(c *fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("user_id").(string)

	var complaints []models.Complaint
	query := config.DB.
		Preload("User").
		Preload("Student", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User")
		}).
		Preload("Attachments").
		Preload("Timeline").
		Order("created_at DESC")

	err := query.Find(&complaints).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to fetch complaints",
			"details": err.Error(),
		})
	}

	// Students only see their own complaints
	if role == "student" && userID != "" {
		filtered := []models.Complaint{}
		for _, c := range complaints {
			if c.UserID.String() == userID {
				filtered = append(filtered, c)
			}
		}
		complaints = filtered
	}

	if len(complaints) == 0 {
		return c.JSON(fiber.Map{"message": "No complaints found", "data": []models.Complaint{}})
	}

	return c.JSON(fiber.Map{"count": len(complaints), "data": complaints})
}

// 🧑‍💼 ADMIN — Update Complaint Status
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

// 🧑‍💼 ADMIN — Delete Complaint
func DeleteComplaint(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := config.DB.Delete(&models.Complaint{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete complaint"})
	}
	return c.JSON(fiber.Map{"message": "Complaint deleted successfully"})
}

// 📊 Filter Complaints by Type (Optional)
func GetComplaintsByType(c *fiber.Ctx) error {
	var complaints []models.Complaint
	complaintType := c.Query("type")

	query := config.DB.Model(&models.Complaint{})
	if complaintType != "" {
		query = query.Where("type = ?", complaintType)
	}

	if err := query.Order("created_at desc").Find(&complaints).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch complaints"})
	}

	return c.JSON(complaints)
}

// 📋 Get Complaint by ID (with full details)
func GetComplaintbyID(c *fiber.Ctx) error {
	complaintID := c.Params("id")
	var complaint models.Complaint

	err := config.DB.
		Preload("User").
		Preload("Student", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User")
		}).
		Preload("Timeline").
		Preload("Attachments").
		First(&complaint, "id = ?", complaintID).Error

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error":   "Complaint not found",
			"details": err.Error(),
		})
	}

	// Fetch user's past complaints (excluding this one)
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
		"attachments":     complaint.Attachments,
		"timeline":        complaint.Timeline,
		"past_complaints": pastComplaints,
	}

	return c.JSON(response)
}

// get all complaints by admin
// 🧑‍💼 ADMIN — Get All Complaints (with optional filters)
func GetAllComplaintsAdmin(c *fiber.Ctx) error {
	var complaints []models.Complaint

	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("user_id").(string)

	query := config.DB.Preload("User").Preload("Student").Preload("Attachments").Preload("Timeline")

	// Optional filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if complaintType := c.Query("type"); complaintType != "" {
		query = query.Where("type = ?", complaintType)
	}

	// If the requester is an admin (not chief_admin), restrict to their block
	if role == string(models.Admin) && userID != "" {
		// get requesting user's block
		var reqUser models.User
		if err := config.DB.First(&reqUser, "id = ?", userID).Error; err == nil {
			if reqUser.Block != "" {
				// complaints are associated to students who have Hostel field; filter by that
				query = query.Joins("JOIN student_models ON student_models.user_id = complaints.user_id").Where("student_models.hostel = ?", reqUser.Block)
			}
		}
	}

	if err := query.Order("created_at desc").Find(&complaints).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch complaints"})
	}

	return c.JSON(fiber.Map{
		"count": len(complaints),
		"data":  complaints,
	})
}
