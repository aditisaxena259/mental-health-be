package models

import (
    "github.com/google/uuid"
)

type StudentModel struct {
    ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
    UserID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`

    Hostel   string    `gorm:"type:text" json:"hostel"`
    RoomNo   string    `gorm:"type:text" json:"room_no"`

    User     User      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;" json:"user"`
}
