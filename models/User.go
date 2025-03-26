package models

import (
    "github.com/google/uuid"
    "time"
)

type RoleType string

const (
    Student   RoleType = "student"
    Admin     RoleType = "admin"
    Counselor RoleType = "counselor"
)

type User struct {
    ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"` // ✅ Changed to `gen_random_uuid()`
    Name      string    `gorm:"not null"`
    Email     string    `gorm:"unique;not null"`
    Password  string    `gorm:"not null"`
    Role      RoleType  `gorm:"type:user_role;not null"`
    CreatedAt time.Time
}
