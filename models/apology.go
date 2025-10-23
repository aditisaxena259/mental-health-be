package models

import (
	"time"

	"github.com/google/uuid"
)

// Optional: Custom status type
type ApologyStatus string

const (
	ApologySubmitted ApologyStatus = "submitted"
	ApologyReviewed  ApologyStatus = "reviewed"
	ApologyAccepted  ApologyStatus = "accepted"
	ApologyRejected  ApologyStatus = "rejected"
)

type ApologyType string
const(
	Apologyforouting ApologyType ="outing"
	Apologyformisconduct ApologyType="misconduct"
	Apologyforotherreason ApologyType="miscellaneous"
)
type Apology struct {
	ID          uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	StudentID   uuid.UUID     `gorm:"type:uuid;not null" json:"student_id"` // use uuid not string
	ApologyType ApologyType   `gorm:"type:text;not null" json:"type"`
	Message     string        `gorm:"type:text;not null"`
	Description string        `gorm:"type:text"`
	Status      ApologyStatus `gorm:"type:text;default:'submitted'" json:"status"`
	Comment     string        `gorm:"type:text"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"created_at"`

	Student StudentModel `gorm:"foreignKey:StudentID;references:UserID" json:"student"`
}

