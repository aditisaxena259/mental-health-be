package models

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken stores tokens for password reset flows
type PasswordResetToken struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"type:text;not null;uniqueIndex" json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// CounselorSlot represents available time slots a counselor offers
type CounselorSlot struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CounselorID uuid.UUID `gorm:"type:uuid;not null;index" json:"counselor_id"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	IsBooked    bool      `gorm:"default:false" json:"is_booked"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// CounselingSession links a student to a counselor slot and stores notes/progress
// CounselingSession is defined in models/Csession.go to preserve existing schema.
