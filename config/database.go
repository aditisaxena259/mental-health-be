package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase connects to PostgreSQL using DATABASE_URL from .env
// Includes safe connection pooling, disabled prepared statements, and extension setup.
func ConnectDatabase() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("❌ DATABASE_URL is not set in environment")
	}

	// ✅ Disable prepared statements (avoids SQLSTATE 08P01)
	// ✅ Safe for all environments (especially Supabase / pgBouncer)
	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // 🔥 disables prepared statements
	}), &gorm.Config{})

	if err != nil {
		return fmt.Errorf("❌ failed to connect to database: %w", err)
	}

	log.Println("✅ Connected to Supabase PostgreSQL (simple protocol enabled)!")

	// Enable UUID extensions (safe to run multiple times)
	if err := database.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		log.Printf("⚠️ Could not ensure uuid-ossp extension: %v", err)
	}
	if err := database.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`).Error; err != nil {
		log.Printf("⚠️ Could not ensure pgcrypto extension: %v", err)
	}

	// ✅ Connection pooling — prevents stale or idle session buildup
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("❌ failed to get database instance: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)               // Maximum open connections
	sqlDB.SetMaxIdleConns(5)                // Keep a few idle
	sqlDB.SetConnMaxLifetime(1 * time.Hour) // Recycle after 1 hour

	DB = database
	return nil
}
