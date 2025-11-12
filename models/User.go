package models

import (
	"time"

	"github.com/google/uuid"
)

type RoleType string

const (
	Student    RoleType = "student"
	Admin      RoleType = "admin"
	ChiefAdmin RoleType = "chief_admin"
	Counselor  RoleType = "counselor"
)

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name     string    `gorm:"not null" json:"name"`
	Email    string    `gorm:"unique;not null" json:"email"`
	Password string    `gorm:"not null" json:"password"`
	Role     RoleType  `gorm:"type:user_role;not null" json:"role"`
	// Block is used to map admin users to a hostel block. For students, block info is in StudentModel.
	Block     string    `gorm:"type:char(1);not null;check:block ~ '^[A-Z]$'" json:"block"`
	CreatedAt time.Time `json:"created_at"`
}
