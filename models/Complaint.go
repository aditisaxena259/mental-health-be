package models

import (
	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/google/uuid"
	"time"
)

type ComplaintType string

const (
	Roommate    ComplaintType = "roommate"
	Plumbing    ComplaintType = "plumbing"
	Cleanliness ComplaintType = "cleanliness"
	Miscellaneous ComplaintType = "miscellaneous"
)

type StatusType1 string

const (
	Open       StatusType = "open"
	InProgress StatusType = "inprogress"
	Resolved   StatusType = "resolved"
)

type Complaint struct {
	ID          uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      uuid.UUID     `gorm:"not null"`
	Type        ComplaintType `gorm:"type:complaint_type;not null"` // Use ENUM
	Description string        `gorm:"type:text;not null"`
	Status      StatusType    `gorm:"type:status_type;not null;default:'open'"`
	CreatedAt   time.Time
}

// ✅ Ensure ENUMs exist before creating the table
func MigrateDatabase() {
	config.DB.Exec(`DO $$ BEGIN 
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'complaint_type') THEN 
            CREATE TYPE complaint_type AS ENUM ('roommate', 'plumbing', 'cleanliness', 'miscellaneous'); 
        END IF; 
    END $$;`)

	config.DB.Exec(`DO $$ BEGIN 
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'status_type') THEN 
            CREATE TYPE status_type AS ENUM ('open', 'inprogress', 'resolved'); 
        END IF; 
    END $$;`)

	config.DB.AutoMigrate(&Complaint{})
}
