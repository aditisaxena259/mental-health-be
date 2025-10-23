package models

import (
    "github.com/google/uuid"
)

type StudentModel struct {
    ID       uuid.UUID `gorm:"type:uuid;primaryKey"`  // Same as User.ID
    UserID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`

    Hostel   string    `json:"hostel"`
    RoomNo   string    `json:"room_no"`

    User     User      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;" json:"user"`
}
