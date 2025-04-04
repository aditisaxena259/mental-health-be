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
    ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
    Name      string    `gorm:"not null" json:"name"`
    Email     string    `gorm:"unique;not null" json:"email"`
    Password  string    `gorm:"not null" json:"password"`
    Role      RoleType  `gorm:"type:user_role;not null" json:"role"`
    CreatedAt time.Time `json:"created_at"`
}

