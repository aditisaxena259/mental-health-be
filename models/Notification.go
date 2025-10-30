package models

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents an in-app notification for a user (admin or student)
type Notification struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Title       string     `gorm:"type:text;not null" json:"title"`
	Message     string     `gorm:"type:text;not null" json:"message"`
	Type        string     `gorm:"type:text;default:'info'" json:"type"` // info|success|warning|error
	RelatedID   *uuid.UUID `gorm:"type:uuid;index" json:"related_id,omitempty"`
	RelatedType *string    `gorm:"type:text" json:"related_type,omitempty"`
	IsRead      bool       `gorm:"default:false" json:"is_read"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}
