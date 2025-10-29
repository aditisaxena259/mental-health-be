package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// POST /api/forgot-password
func ForgotPassword(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}
	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// Don't reveal whether email exists
		return c.JSON(fiber.Map{"message": "If the email exists, a reset token has been sent"})
	}

	// generate token
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)

	pr := models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if err := config.DB.Create(&pr).Error; err != nil {
		// log and return generic message
		return c.JSON(fiber.Map{"message": "If the email exists, a reset token has been sent"})
	}

	// In DEV_MODE, return the token in the response so tests can pick it up.
	if os.Getenv("DEV_MODE") == "true" {
		return c.JSON(fiber.Map{"message": "If the email exists, a reset token has been sent", "token": pr.Token})
	}

	// In production you'd send the token via email. For now we simply return success.
	return c.JSON(fiber.Map{"message": "If the email exists, a reset token has been sent"})
}

// POST /api/reset-password
func ResetPassword(c *fiber.Ctx) error {
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}
	var pr models.PasswordResetToken
	if err := config.DB.Where("token = ?", input.Token).First(&pr).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid token"})
	}
	if time.Now().After(pr.ExpiresAt) {
		return c.Status(400).JSON(fiber.Map{"error": "token expired"})
	}
	var user models.User
	if err := config.DB.First(&user, "id = ?", pr.UserID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "user not found"})
	}
	// hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
	user.Password = string(hashedPassword)
	config.DB.Save(&user)
	// delete token
	config.DB.Delete(&pr)
	return c.JSON(fiber.Map{"message": "Password reset successfully"})
}

// GET /api/counselors/:id/slots  -- list available slots for counselor
func ListCounselorSlots(c *fiber.Ctx) error {
	cid := c.Params("id")
	var slots []models.CounselorSlot
	config.DB.Where("counselor_id = ? AND is_booked = false", cid).Find(&slots)
	return c.JSON(slots)
}

// POST /api/counselors/:id/slots - create a slot (counselor or admin can create)
func CreateCounselorSlot(c *fiber.Ctx) error {
	cid := c.Params("id")
	var input struct {
		Start string `json:"start"` // RFC3339
		End   string `json:"end"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}
	start, err := time.Parse(time.RFC3339, input.Start)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid start time, use RFC3339"})
	}
	end, err := time.Parse(time.RFC3339, input.End)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid end time, use RFC3339"})
	}
	slot := models.CounselorSlot{
		ID:          uuid.New(),
		CounselorID: uuid.MustParse(cid),
		Start:       start,
		End:         end,
	}
	if err := config.DB.Create(&slot).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create slot"})
	}
	return c.JSON(fiber.Map{"message": "slot created", "slot": slot})
}

// DEV: GET /api/dev/reset-token?email=...  - returns latest reset token for email when DEV_MODE=true
func DevGetResetToken(c *fiber.Ctx) error {
	if os.Getenv("DEV_MODE") != "true" {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	email := c.Query("email")
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	var pr models.PasswordResetToken
	if err := config.DB.Where("user_id = ?", user.ID).Order("created_at desc").First(&pr).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "token not found"})
	}
	return c.JSON(fiber.Map{"token": pr.Token})
}

// POST /api/admin/counselors/:id/book - admin books a slot for a student
func BookCounselorSlot(c *fiber.Ctx) error {
	counselorID := c.Params("id")
	var input struct {
		SlotID            string `json:"slot_id"`
		StudentIdentifier string `json:"student_identifier"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}
	// find slot
	var slot models.CounselorSlot
	if err := config.DB.First(&slot, "id = ? AND counselor_id = ?", input.SlotID, counselorID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "slot not found"})
	}
	if slot.IsBooked {
		return c.Status(400).JSON(fiber.Map{"error": "slot already booked"})
	}
	// find student user by student identifier
	var student models.StudentModel
	if err := config.DB.Where("student_identifier = ?", input.StudentIdentifier).First(&student).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}
	// create counseling session
	cs := models.CounselingSession{
		ID:          uuid.New(),
		StudentID:   student.UserID,
		CounselorID: uuid.MustParse(counselorID),
		StartTime:   slot.Start,
		EndTime:     slot.End,
		Status:      models.Pending,
	}
	// mark slot booked
	slot.IsBooked = true
	tx := config.DB.Begin()
	tx.Save(&slot)
	tx.Create(&cs)
	tx.Commit()
	return c.JSON(fiber.Map{"message": "Slot booked", "session": cs})
}

// POST /api/counselor/sessions/:id/update - counselor updates notes/progress
func CounselorUpdateSession(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	var input struct {
		Notes    string            `json:"notes"`
		Progress string            `json:"progress"`
		Status   models.StatusType `json:"status"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}
	var session models.CounselingSession
	if err := config.DB.First(&session, "id = ?", sessionID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "session not found"})
	}
	// Notes and Progress fields exist on the model; update if provided
	if input.Notes != "" {
		session.Notes = input.Notes
	}
	if input.Progress != "" {
		session.Progress = input.Progress
	}
	if input.Status != "" {
		session.Status = input.Status
	}
	config.DB.Save(&session)
	return c.JSON(fiber.Map{"message": "Session updated", "session": session})
}

// GET /api/student/profile - student views own profile
func GetOwnProfile(c *fiber.Ctx) error {
	uid, _ := c.Locals("user_id").(string)
	if uid == "" {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var user models.User
	if err := config.DB.First(&user, "id = ?", uid).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	var student models.StudentModel
	config.DB.Where("user_id = ?", uid).Preload("User").First(&student)
	profile := fiber.Map{
		"id":                 user.ID,
		"name":               user.Name,
		"email":              user.Email,
		"block":              user.Block,
		"hostel":             student.Hostel,
		"room_no":            student.RoomNo,
		"student_identifier": student.StudentIdentifier,
	}
	return c.JSON(profile)
}

// GET /api/admin/student/:student_identifier - admin views student profile by student identifier
func GetStudentProfileAdmin(c *fiber.Ctx) error {
	sid := c.Params("student_identifier")
	var student models.StudentModel
	if err := config.DB.Where("student_identifier = ?", sid).Preload("User").First(&student).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}
	profile := fiber.Map{
		"id":                 student.User.ID,
		"name":               student.User.Name,
		"email":              student.User.Email,
		"block":              student.User.Block,
		"hostel":             student.Hostel,
		"room_no":            student.RoomNo,
		"student_identifier": student.StudentIdentifier,
	}
	return c.JSON(profile)
}
