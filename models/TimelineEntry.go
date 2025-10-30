package models

import (
	"time"

	"github.com/google/uuid"
)

type TimelineEntry struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ComplaintID uuid.UUID `gorm:"not null"`
	Author      string
	Message     string
	Timestamp   time.Time
}

func (TimelineEntry) TableName() string {
	return "timeline_entries"
}
