package models

import (
	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/google/uuid"
	"time"
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

type Complaint struct {
	ID          uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title       string          `gorm:"type:text;not null"`
	UserID      uuid.UUID       `gorm:"type:uuid;not null"`
	Type        ComplaintType   `gorm:"type:complaint_type;not null"`
	Description string          `gorm:"type:text;not null"`
	Status      ComplaintStatus `gorm:"type:status_type;default:'open'"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"`

	User        User            `gorm:"foreignKey:UserID;references:ID" json:"user"`
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
