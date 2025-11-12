package main

import (
	"fmt"
	"log"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/aditisaxena259/mental-health-be/models"
	"github.com/google/uuid"
)

func main() {
	// Initialize database
	config.ConnectDatabase()

	// Find all student users without student_models
	var users []models.User
	config.DB.Where("role = ?", "student").Find(&users)

	for _, user := range users {
		// Check if student_model already exists
		var existing models.StudentModel
		err := config.DB.Where("user_id = ?", user.ID).First(&existing).Error
		if err == nil {
			log.Printf("✓ Student model already exists for %s\n", user.Email)
			continue
		}

		// Create student_model
		studentID := "STU-001"
		roomNo := "101"
		if user.Email == "student2@uni.com" {
			studentID = "STU-002"
			roomNo = "102"
		}

		student := models.StudentModel{
			ID:                uuid.New(),
			UserID:            user.ID,
			StudentIdentifier: studentID,
			Block:             user.Block,
			RoomNo:            roomNo,
		}

		if err := config.DB.Create(&student).Error; err != nil {
			log.Printf("✗ Failed to create student model for %s: %v\n", user.Email, err)
		} else {
			log.Printf("✓ Created student model for %s (Block: %s, Room: %s)\n", user.Email, student.Block, student.RoomNo)
		}
	}

	fmt.Println("\n✅ Student models fix complete!")
}
