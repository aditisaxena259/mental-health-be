package models

import (
	"fmt"
	"log"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func SeedData() {
	var count int64
	config.DB.Model(&User{}).Count(&count)
	if count > 0 {
		log.Println("🌱 Seed skipped (users already exist)")
		return
	}

	// Admin
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 14)
	admin := User{
		ID:       uuid.New(),
		Name:     "Warden Admin",
		Email:    "admin@hostel.com",
		Password: string(hash),
		Role:     "admin",
	}
	config.DB.Create(&admin)

	// Students
	for i := 1; i <= 2; i++ {
		sid := uuid.New()
		shash, _ := bcrypt.GenerateFromPassword([]byte("student123"), 14)
		user := User{
			ID:       sid,
			Name:     fmt.Sprintf("Student%d", i),
			Email:    fmt.Sprintf("student%d@uni.com", i),
			Password: string(shash),
			Role:     "student",
		}
		config.DB.Create(&user)
		student := StudentModel{
			UserID: user.ID,
			Hostel: "A-Block",
			RoomNo: fmt.Sprintf("10%d", i),
		}
		config.DB.Create(&student)
	}

	log.Println("🌱 Seeded admin and sample students!")
}
