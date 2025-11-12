package controllers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/helpers"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// 🧑‍🎓 STUDENT — Submit Apology Letter
func SubmitApology(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: missing user ID"})
	}

	var input struct {
		Type        models.ApologyType `form:"type" json:"type"`
		Message     string             `form:"message" json:"message"`
		Description string             `form:"description" json:"description"`
	}
	if form, _ := c.MultipartForm(); form != nil {
		input.Type = models.ApologyType(c.FormValue("type"))
		input.Message = c.FormValue("message")
		input.Description = c.FormValue("description")
	} else if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input", "details": err.Error()})
	}

	if input.Message == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Message field is required"})
	}

	studentUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	apology := models.Apology{
		StudentID:   studentUUID,
		ApologyType: input.Type,
		Message:     input.Message,
		Description: input.Description,
		Status:      models.ApologySubmitted,
	}

	// set StudentIdentifier if available
	var sm models.StudentModel
	if err := config.DB.Where("user_id = ?", studentUUID).First(&sm).Error; err == nil {
		// use StudentIdentifier for external mapping
		// Apology model will include StudentIdentifier in JSON response if present
		// but keep StudentID for DB relations
		apology.StudentID = studentUUID
	}

	if err := config.DB.Create(&apology).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to submit apology", "details": err.Error()})
	}

	// If multipart attachments are included, upload to Cloudinary and save records
	if form, _ := c.MultipartForm(); form != nil {
		files := form.File["attachments"]
		if len(files) > 0 {
			cld, cldErr := helpers.InitCloudinary()
			// Ensure temp dir exists
			_ = os.MkdirAll("./uploads/apologies", 0755)
			for _, fh := range files {
				if fh == nil {
					continue
				}
				// Only accept JPEGs
				ext := filepath.Ext(fh.Filename)
				if ext == "" {
					ext = ".jpg"
				}
				tmpPath := fmt.Sprintf("./uploads/apologies/%s_%s", apology.ID.String(), fh.Filename)
				if err := c.SaveFile(fh, tmpPath); err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "Failed to save attachment", "details": err.Error()})
				}
				// default values for fallback
				var fileURL, publicID, sizeStr string
				// try Cloudinary if configured
				if cldErr == nil {
					if uploadRes, upErr := cld.UploadJPEG(tmpPath, "apologies/"+apology.ID.String(), uuid.New().String()); upErr == nil {
						fileURL = uploadRes.SecureURL
						publicID = uploadRes.PublicID
						sizeStr = fmt.Sprintf("%d", uploadRes.Bytes)
						// cleanup tmp after successful upload
						_ = os.Remove(tmpPath)
					} else {
						// fallback to local file
						if fi, err := os.Stat(tmpPath); err == nil {
							sizeStr = fmt.Sprintf("%d", fi.Size())
						}
						fileURL = tmpPath
						publicID = ""
					}
				} else {
					// no Cloudinary configured; fallback to local file path
					if fi, err := os.Stat(tmpPath); err == nil {
						sizeStr = fmt.Sprintf("%d", fi.Size())
					}
					fileURL = tmpPath
					publicID = ""
				}
				att := models.ApologyAttachment{
					ID:        uuid.New(),
					ApologyID: apology.ID,
					FileName:  fh.Filename,
					FileURL:   fileURL,
					PublicID:  publicID,
					Size:      sizeStr,
				}
				if err := config.DB.Create(&att).Error; err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "Failed to persist attachment", "details": err.Error()})
				}
			}
		}
	}

	// create notifications for admins of the student's block + chief admins
	go func(a models.Apology) {
		var sm models.StudentModel
		if err := config.DB.Where("user_id = ?", a.StudentID).First(&sm).Error; err != nil {
			return
		}
		hostel := sm.Hostel
		var admins []models.User
		if hostel != "" {
			config.DB.Where("role = ? AND block = ?", models.Admin, hostel).Find(&admins)
		}
		var chiefs []models.User
		config.DB.Where("role = ?", models.ChiefAdmin).Find(&chiefs)
		// deduplicate
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
		related := a.ID
		rtype := "apology"
		for _, adm := range admins {
			n := models.Notification{
				ID:          uuid.New(),
				UserID:      adm.ID,
				Title:       "New Apology Submitted",
				Message:     "A student has submitted an apology: " + a.Message,
				Type:        "info",
				RelatedID:   &related,
				RelatedType: &rtype,
			}
			config.DB.Create(&n)
		}
	}(apology)

	// ✅ Preload after creation so response includes Student details and attachments
	config.DB.Preload("Student.User").Preload("Attachments").First(&apology, "id = ?", apology.ID)

	return c.JSON(fiber.Map{
		"message": "Apology letter submitted successfully",
		"data":    apology,
	})
}

// 🧑‍🎓 STUDENT — Get Own Apologies
func GetStudentApologies(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var apologies []models.Apology
	if err := config.DB.
		Preload("Student.User").Preload("Attachments").
		Where("student_id = ?", userID).
		Order("created_at desc").
		Find(&apologies).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch apologies"})
	}

	if len(apologies) == 0 {
		return c.JSON(fiber.Map{"message": "No apologies found", "data": []models.Apology{}})
	}

	return c.JSON(fiber.Map{"count": len(apologies), "data": apologies})
}

// 🧑‍💼 ADMIN — Get All or Filtered Apologies
func GetApologies(c *fiber.Ctx) error {
	var apologies []models.Apology
	query := config.DB.Preload("Student.User").Preload("Attachments")

	if apologyType := c.Query("type"); apologyType != "" {
		query = query.Where("apology_type = ?", apologyType)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at desc").Find(&apologies).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch apologies"})
	}

	return c.JSON(fiber.Map{"count": len(apologies), "data": apologies})
}

// 🧑‍💼 ADMIN — Get Apology by ID
func GetApologyByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var apology models.Apology

	if err := config.DB.Preload("Student.User").Preload("Attachments").First(&apology, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Apology not found"})
	}

	return c.JSON(apology)
}

// 🧑‍💼 ADMIN — Review or Update Apology Status
func ReviewApology(c *fiber.Ctx) error {
	id := c.Params("id")

	var input struct {
		Status  models.ApologyStatus `json:"status"`
		Comment string               `json:"comment"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var apology models.Apology
	if err := tx.First(&apology, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "Apology not found"})
	}

	apology.Status = input.Status
	apology.Comment = input.Comment

	if err := tx.Save(&apology).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update apology"})
	}

	tx.Commit()

	// ✅ Load Student details for the response
	config.DB.Preload("Student.User").First(&apology, "id = ?", id)

	// Notify student about apology status change (reviewed/accepted/rejected)
	go func(a models.Apology, status models.ApologyStatus) {
		var title, message, ntype string
		switch status {
		case models.ApologySubmitted:
			return
		case models.ApologyReviewed:
			title = "Apology Under Review"
			message = "Your apology letter is being reviewed by the warden."
			ntype = "info"
		case models.ApologyAccepted:
			title = "Apology Accepted"
			message = "Your apology has been accepted."
			ntype = "success"
		case models.ApologyRejected:
			title = "Apology Rejected"
			message = "Your apology has been rejected. Please contact the warden for details."
			ntype = "warning"
		default:
			return
		}

		related := a.ID
		rtype := "apology"
		n := models.Notification{
			ID:          uuid.New(),
			UserID:      a.StudentID,
			Title:       title,
			Message:     message,
			Type:        ntype,
			RelatedID:   &related,
			RelatedType: &rtype,
		}
		config.DB.Create(&n)
	}(apology, input.Status)

	return c.JSON(fiber.Map{
		"message": "Apology reviewed successfully",
		"data":    apology,
	})
}

// 🧾 ADMIN — Pending Count
func GetPendingApology(c *fiber.Ctx) error {
	var count int64
	if err := config.DB.Model(&models.Apology{}).Where("status = ?", models.ApologySubmitted).Count(&count).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count pending apologies"})
	}
	return c.JSON(fiber.Map{"pending_count": count})
}
