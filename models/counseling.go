// File removed

// This file contained the counselor logic which has been removed.

// The following types were previously defined in this file:
// - CounselorSlot
// - CounselingSession
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

// CounselingSession links a student to a counselor slot and stores notes/progress
// CounselingSession is defined in models/Csession.go to preserve existing schema.
