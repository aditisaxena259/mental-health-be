package models

import (
	"time"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/google/uuid"
)

type ComplaintType string

const (
	Roommate      ComplaintType = "roommate"
	Plumbing      ComplaintType = "plumbing"
	Cleanliness   ComplaintType = "cleanliness"
	Electricity   ComplaintType = "electricity"
	LostFound     ComplaintType = "Lost and Found"
	Miscellaneous ComplaintType = "Other Issues"
)

type ComplaintStatus string

const (
	Open       ComplaintStatus = "open"
	InProgress ComplaintStatus = "inprogress"
	Resolved   ComplaintStatus = "resolved"
)

type ComplaintPriority string

const (
	PriorityLow    ComplaintPriority = "low"
	PriorityMedium ComplaintPriority = "medium"
	PriorityHigh   ComplaintPriority = "high"
)

type Complaint struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title  string    `gorm:"type:text;not null"`
	UserID uuid.UUID `gorm:"type:uuid;not null"`
	// StudentIdentifier links the complaint to the external student identifier (e.g. roll number)
	StudentIdentifier string            `gorm:"type:text;index" json:"student_identifier"`
	Type              ComplaintType     `gorm:"type:complaint_type;not null"`
	Description       string            `gorm:"type:text;not null"`
	Priority          ComplaintPriority `gorm:"type:text;default:'medium'" json:"priority"`
	Status            ComplaintStatus   `gorm:"type:status_type;default:'open'"`
	CreatedAt         time.Time         `gorm:"autoCreateTime"`

	User User `gorm:"foreignKey:UserID;references:ID" json:"user"`
	// Fix relationship: UserID (complaint) -> UserID (student_models)
	Student     StudentModel    `gorm:"foreignKey:UserID;references:UserID" json:"student"`
	Attachments []Attachment    `gorm:"foreignKey:ComplaintID" json:"attachments"`
	Timeline    []TimelineEntry `gorm:"foreignKey:ComplaintID" json:"timeline"`
}

func MigrateDatabase() {
	config.DB.Exec(`DO $$ BEGIN 
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'complaint_type') THEN 
            CREATE TYPE complaint_type AS ENUM (
                'roommate', 'plumbing', 'cleanliness', 'electricity', 'Lost and Found', 'Other Issues'
            ); 
        END IF; 
    END $$;`)

	config.DB.Exec(`DO $$ BEGIN 
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'status_type') THEN 
            CREATE TYPE status_type AS ENUM ('open', 'inprogress', 'resolved'); 
        END IF; 
    END $$;`)

	config.DB.AutoMigrate(&User{}, &Attachment{}, &TimelineEntry{}, &Complaint{})
}
