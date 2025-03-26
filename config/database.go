package config

import (
    "fmt"
    "log"
    "os"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() error {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        log.Fatal("DATABASE_URL is not set")
    }

    database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }

    log.Println("Connected to Supabase PostgreSQL!")

    
    database.Exec(`DO $$ BEGIN 
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN 
            CREATE TYPE user_role AS ENUM ('student', 'admin', 'counselor'); 
        END IF; 
    END $$;`)
    database.Exec(`DO $$ BEGIN 
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'complaint_type') THEN 
            CREATE TYPE complaint_type AS ENUM ('roommate', 'plumbing', 'cleanliness', 'miscellaneous'); 
        END IF; 
    END $$;`)

	database.Exec(`DO $$ BEGIN 
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'status_type') THEN 
            CREATE TYPE status_type AS ENUM ('open', 'inprogress', 'resolved'); 
        END IF; 
    END $$;`)

    DB = database
    return nil
}
