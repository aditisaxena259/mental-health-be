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
		&Notification{},
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

	// --- Ensure student_models has student_identifier column and unique index ---
	config.DB.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='student_models' AND column_name='student_identifier') THEN
			ALTER TABLE student_models ADD COLUMN student_identifier text;
		END IF;
		-- Create unique index on student_identifier if not exists
		IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename='student_models' AND indexname='idx_student_models_student_identifier_unique') THEN
			BEGIN
				-- Try to create unique index; if duplicate values exist this will fail and we'll leave the index absent
				EXECUTE 'CREATE UNIQUE INDEX CONCURRENTLY idx_student_models_student_identifier_unique ON student_models (student_identifier)';
			EXCEPTION WHEN others THEN
				-- ignore index creation errors (duplicates, locks)
				NULL;
			END;
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

	// --- Create or migrate notifications table to new schema ---
	config.DB.Exec(`DO $$ BEGIN
		-- If notifications table doesn't exist, create with the new schema
		IF NOT EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = CURRENT_SCHEMA() AND tablename = 'notifications') THEN
			CREATE TABLE notifications (
				id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
				user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				title text NOT NULL,
				message text NOT NULL,
				type text DEFAULT 'info',
				related_id uuid,
				related_type text,
				is_read boolean DEFAULT false,
				created_at timestamptz DEFAULT now(),
				updated_at timestamptz DEFAULT now()
			);
		ELSE
			-- Table exists: ensure new columns exist and migrate legacy columns
			-- Add user_id if missing and copy from admin_id when present
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='notifications' AND column_name='user_id') THEN
				ALTER TABLE notifications ADD COLUMN user_id uuid;
				IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='notifications' AND column_name='admin_id') THEN
					UPDATE notifications SET user_id = admin_id WHERE user_id IS NULL;
				END IF;
			END IF;

			-- Add message column and copy from body if body exists
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='notifications' AND column_name='message') THEN
				ALTER TABLE notifications ADD COLUMN message text;
				IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='notifications' AND column_name='body') THEN
					UPDATE notifications SET message = body WHERE message IS NULL;
				END IF;
			END IF;

			-- Add type, related_id, related_type, updated_at if missing
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='notifications' AND column_name='type') THEN
				ALTER TABLE notifications ADD COLUMN type text DEFAULT 'info';
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='notifications' AND column_name='related_id') THEN
				ALTER TABLE notifications ADD COLUMN related_id uuid;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='notifications' AND column_name='related_type') THEN
				ALTER TABLE notifications ADD COLUMN related_type text;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='notifications' AND column_name='updated_at') THEN
				ALTER TABLE notifications ADD COLUMN updated_at timestamptz DEFAULT now();
			END IF;
		END IF;
	END $$;`)
}
