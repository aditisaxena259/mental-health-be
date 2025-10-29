package controllers

import (
	"strings"
	"time"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/helpers"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Signup(c *fiber.Ctx) error {
	data := make(map[string]string)

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if data["name"] == "" || data["email"] == "" || data["password"] == "" || data["role"] == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name, email, password, and role are required"})
	}

	// Require room and hostel/block depending on role
	role := models.RoleType(data["role"])
	email := data["email"]
	if role == models.Admin || role == models.ChiefAdmin {
		// admin must provide block
		if data["block"] == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Hostel block is required for admin signups"})
		}
		// domain check
		if !strings.HasSuffix(strings.ToLower(email), "@hostel.com") {
			return c.Status(400).JSON(fiber.Map{"error": "Admin signup requires an @hostel.com email"})
		}
	} else if role == models.Student {
		// student must provide hostel and room
		if data["hostel"] == "" || data["room_no"] == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Hostel and room number are required for student signups"})
		}
		if !strings.HasSuffix(strings.ToLower(email), "@uni.com") {
			return c.Status(400).JSON(fiber.Map{"error": "Student signup requires an @uni.com email"})
		}
	}

	// ✅ Check if user already exists
	var existingUser models.User
	err := config.DB.Where("email = ?", data["email"]).First(&existingUser).Error

	if err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "User with this email already exists"})
	} else if err != gorm.ErrRecordNotFound {
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	// ✅ Hash the password safely
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error hashing password"})
	}

	// ✅ Create the new user
	user := models.User{
		ID:       uuid.New(),
		Name:     data["name"],
		Email:    data["email"],
		Password: string(hashedPassword),
		Role:     models.RoleType(data["role"]),
		Block:    data["block"],
	}

	// ✅ Save user to the database
	if err := config.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
	}

	// 🆕 If role = "student", also create a StudentModel record
	if user.Role == models.Student {
		student := models.StudentModel{
			UserID: user.ID,
			Hostel: data["hostel"],
			RoomNo: data["room_no"],
		}
		if err := config.DB.Create(&student).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create student record"})
		}
	}

	// For admins, ensure their block is persisted in User.Block - already set above

	return c.JSON(fiber.Map{"message": "User created successfully"})
}

func Login(c *fiber.Ctx) error {
	var data map[string]string
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	var user models.User
	config.DB.Where("email = ?", data["email"]).First(&user)

	if user.ID == uuid.Nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid email/password"})
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data["password"]))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid email/password"})
	}

	token, err := helpers.GenerateJWT(user.ID.String(), string(user.Role))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	// Set token as HTTP-only cookie
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour), // 1-day expiration
		HTTPOnly: true,
		Secure:   true, // Enable for HTTPS
		SameSite: "Strict",
	})

	// ✅ Return role in JSON response
	return c.JSON(fiber.Map{
		"message": "Login successful",
		"token":   token,
		"role":    user.Role, // Include role in the response
	})
}

func Logout(c *fiber.Ctx) error {

	c.Cookie(&fiber.Cookie{
		Name:     "token", // Replace with your actual cookie name
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Expire immediately
		HTTPOnly: true,
	})

	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}
