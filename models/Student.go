package models

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type StudentModel struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`

	// StudentIdentifier is the externally-provided student id (e.g. university roll number)
	// This will be used to map complaints/apologies and other student-specific records.
	// StudentIdentifier is stored in DB as `student_identifier` (snake_case by GORM).
	// Keep the JSON field as `student_id` for API compatibility.
	StudentIdentifier string `gorm:"type:text;uniqueIndex;not null" json:"student_id"`

	Block  string `gorm:"type:char(1);not null;check:block ~ '^[A-Z]$'" json:"block"`
	Hostel string `gorm:"-" json:"hostel"` // Always same as Block, for API alias
	RoomNo string `gorm:"type:text" json:"room_no"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;" json:"user"`
}

// MarshalJSON to always include hostel as alias for block
func (s StudentModel) MarshalJSON() ([]byte, error) {
	type Alias StudentModel
	return []byte(fmt.Sprintf(`{"id":"%s","user_id":"%s","student_id":"%s","block":"%s","hostel":"%s","room_no":"%s","user":%s}`,
		s.ID, s.UserID, s.StudentIdentifier, s.Block, s.Block, s.RoomNo, toJSON(s.User))), nil
}

func toJSON(u User) string {
	b, _ := json.Marshal(u)
	return string(b)
}
