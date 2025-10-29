package models

import "github.com/aditisaxena259/mental-health-be/config"

// AutoMigrateAll sets up all ENUMs, adds missing values if needed,
// and migrates all tables in proper dependency order.
func AutoMigrateAll() {
	config.DB.Exec(`
		DO $$ BEGIN 
			-- User roles
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN 
				CREATE TYPE user_role AS ENUM ('student', 'admin', 'chief_admin', 'counselor'); 
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
			-- Apology types
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'apology_type') THEN 
				CREATE TYPE apology_type AS ENUM (
					'outing',
					'misconduct',
					'miscellaneous'
				);
			END IF;

			-- Apology statuses
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'apology_status') THEN 
				CREATE TYPE apology_status AS ENUM (
					'submitted', 
					'reviewed', 
					'accepted', 
					'rejected'
				);
			END IF;
		END $$;
	`)

	// --- Ensure missing ENUM values exist ---
	config.DB.Exec(`ALTER TYPE complaint_type ADD VALUE IF NOT EXISTS 'electricity';`)
	config.DB.Exec(`ALTER TYPE complaint_type ADD VALUE IF NOT EXISTS 'Lost and Found';`)
	config.DB.Exec(`ALTER TYPE complaint_type ADD VALUE IF NOT EXISTS 'Other Issues';`)

	// Ensure user_role enum has chief_admin for existing DBs
	config.DB.Exec(`ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'chief_admin';`)

	// --- Migrate all tables in dependency order ---
	config.DB.AutoMigrate(
		&User{},
		&StudentModel{},
		&Complaint{},
		&Attachment{},
		&TimelineEntry{},
		&Apology{}, // ✅ only this line added
		&PasswordResetToken{},
		&CounselorSlot{},
	)

	// --- Enforce correct column types ---
	config.DB.Exec(`
		ALTER TABLE complaints 
		ALTER COLUMN type TYPE complaint_type USING type::complaint_type,
		ALTER COLUMN status TYPE status_type USING status::status_type;
	`)

	// --- Add missing columns that may not exist in older DBs ---
	config.DB.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='complaints' AND column_name='student_identifier') THEN
			ALTER TABLE complaints ADD COLUMN student_identifier text;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='apologies' AND column_name='student_identifier') THEN
			ALTER TABLE apologies ADD COLUMN student_identifier text;
		END IF;
		-- Add priority column for complaints (text with default 'medium')
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='complaints' AND column_name='priority') THEN
			ALTER TABLE complaints ADD COLUMN priority text DEFAULT 'medium';
		END IF;
	END $$;`)

	// --- Create password_reset_tokens table if not present (GORM sometimes misses creation in edge cases) ---
	config.DB.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = CURRENT_SCHEMA() AND tablename = 'password_reset_tokens') THEN
			CREATE TABLE password_reset_tokens (
				id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				token text NOT NULL UNIQUE,
				expires_at timestamptz,
				created_at timestamptz DEFAULT now()
			);
		END IF;
	END $$;`)
}
