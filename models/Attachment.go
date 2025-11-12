package models

import "github.com/google/uuid"

type Attachment struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ComplaintID uuid.UUID `gorm:"not null"`
	FileName    string
	FileURL     string
	PublicID    string // cloudinary public id
	Size        string
	FilePath    string
}

func (Attachment) TableName() string {
	return "attachments"
}
