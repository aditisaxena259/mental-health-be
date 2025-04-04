package models

import (
	"time"
)

// Optional: Custom status type
type ApologyStatus string

const (
	ApologySubmitted ApologyStatus = "submitted"
	ApologyReviewed  ApologyStatus = "reviewed"
	ApologyAccepted  ApologyStatus = "accepted"
	ApologyRejected  ApologyStatus = "rejected"
)

type ApologyLetter struct {
	ID        uint          `json:"id" gorm:"primaryKey"`
	StudentID string        `json:"student_id"`                     // Foreign key / reference
	Message   string        `json:"message"`                        // Apology content
	Status    ApologyStatus `json:"status" gorm:"default:submitted"`// Current status
	CreatedAt time.Time     `json:"created_at"`                     // Timestamp
}
