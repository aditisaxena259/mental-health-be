package models

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents an in-app notification for admins
type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AdminID   uuid.UUID `gorm:"type:uuid;not null;index" json:"admin_id"`
	Title     string    `gorm:"type:text;not null" json:"title"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	Link      string    `gorm:"type:text" json:"link"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
