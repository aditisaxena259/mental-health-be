package models

import "github.com/aditisaxena259/mental-health-be/config"

// AutoMigrateAll sets up all ENUMs, adds missing values if needed,
// and migrates all tables in proper dependency order.
func AutoMigrateAll() {
	config.DB.Exec(`
		DO $$ BEGIN 
			-- User roles
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN 
				CREATE TYPE user_role AS ENUM ('student', 'admin', 'counselor'); 
			END IF;

			-- Complaint types
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'complaint_type') THEN 
				CREATE TYPE complaint_type AS ENUM (
					'roommate', 
					'plumbing', 
					'cleanliness', 
					'electricity', 
					'Lost and Found', 
					'Other Issues'
				); 
			END IF;

			-- Complaint statuses
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'status_type') THEN 
				CREATE TYPE status_type AS ENUM (
					'open', 
					'inprogress', 
					'resolved'
				); 
			END IF;
		END $$;
	`)

	// --- Ensure missing ENUM values exist ---
	config.DB.Exec(`ALTER TYPE complaint_type ADD VALUE IF NOT EXISTS 'electricity';`)
	config.DB.Exec(`ALTER TYPE complaint_type ADD VALUE IF NOT EXISTS 'Lost and Found';`)
	config.DB.Exec(`ALTER TYPE complaint_type ADD VALUE IF NOT EXISTS 'Other Issues';`)

	// --- Migrate all tables in dependency order ---
	config.DB.AutoMigrate(
		&User{},
		&StudentModel{},
		&Complaint{},
		&Attachment{},
		&TimelineEntry{},
		&Apology{},
	)

	// --- Enforce correct column types ---
	config.DB.Exec(`
		ALTER TABLE complaints 
		ALTER COLUMN type TYPE complaint_type USING type::complaint_type,
		ALTER COLUMN status TYPE status_type USING status::status_type;
	`)
}
