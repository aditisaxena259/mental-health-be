package models

import (
	"time"

	"github.com/aditisaxena259/mental-health-be/config"
	"github.com/google/uuid"
)

type StatusType string

const (
	Pending   StatusType = "pending"
	Confirmed StatusType = "confirmed"
	Completed StatusType = "completed"
	Cancelled StatusType = "cancelled"
)

type CounselingSession struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primarykey"`
	StudentID   uuid.UUID  `gorm:"not null"`
	CounselorID uuid.UUID  `gorm:"not null"`
	StartTime   time.Time  `gorm:"not null"`
	EndTime     time.Time  `gorm:"not null"`
	Status      StatusType `gorm:"type:status_type;not null;default:'pending'"`
	Notes       string     `gorm:"type:text" json:"notes"`
	Progress    string     `gorm:"type:text" json:"progress"`
}

// Migrate the database
func MigrateDatabase1() {
	config.DB.Exec(`DO $$ BEGIN 
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'status_type') THEN 
            CREATE TYPE status_type AS ENUM ('pending', 'confirmed', 'completed', 'cancelled'); 
        END IF; 
    END $$;`)

	config.DB.AutoMigrate(&CounselingSession{})

	config.DB.Exec("ALTER TABLE counseling_sessions ADD CONSTRAINT IF NOT EXISTS unique_booking UNIQUE (counselor_id, start_time, end_time);")

	// Add columns if they don't exist (safe for existing DBs)
	config.DB.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='counseling_sessions' AND column_name='notes') THEN
			ALTER TABLE counseling_sessions ADD COLUMN notes text;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='counseling_sessions' AND column_name='progress') THEN
			ALTER TABLE counseling_sessions ADD COLUMN progress text;
		END IF;
	END $$;`)
}
