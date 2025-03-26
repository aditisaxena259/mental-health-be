package models

import (
	"time"

	"github.com/google/uuid"
)

type CounselorAvailability struct{
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CounselorID uuid.UUID `gorm:"type:uuid;not null"`
	Date time.Time `gorm:"not null"`
	StartTime time.Time `gorm:"not null"`
	EndTime time.Time `gorm:"not null"`
	IsBooked bool `gorm:"not null;default:false"`
	
}