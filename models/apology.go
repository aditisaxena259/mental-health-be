package models

import (
	"time"

	"github.com/google/uuid"
)

type ApologyStatus string

const (
	ApologySubmitted ApologyStatus = "submitted"
	ApologyReviewed  ApologyStatus = "reviewed"
	ApologyAccepted  ApologyStatus = "accepted"
	ApologyRejected  ApologyStatus = "rejected"
)

type ApologyType string

const (
	ApologyForOuting      ApologyType = "outing"
	ApologyForMisconduct  ApologyType = "misconduct"
	ApologyForOtherReason ApologyType = "miscellaneous"
)

type Apology struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	StudentID uuid.UUID `gorm:"type:uuid;not null" json:"student_id"`
	// External student identifier (e.g., roll number)
	StudentIdentifier string        `gorm:"type:text;index" json:"student_identifier"`
	ApologyType       ApologyType   `gorm:"type:text;not null" json:"type"`
	Message           string        `gorm:"type:text;not null" json:"message"`
	Description       string        `gorm:"type:text" json:"description"`
	Status            ApologyStatus `gorm:"type:text;default:'submitted'" json:"status"`
	Comment           string        `gorm:"type:text" json:"comment"`
	CreatedAt         time.Time     `gorm:"autoCreateTime" json:"created_at"`

	// Fix relationship: StudentID (apology) -> ID (student_models)
	Student StudentModel `gorm:"foreignKey:StudentID;references:UserID;constraint:OnDelete:CASCADE;" json:"student"`

	// Attachments uploaded with the apology
	Attachments []ApologyAttachment `gorm:"foreignKey:ApologyID;constraint:OnDelete:CASCADE;" json:"attachments"`
}
