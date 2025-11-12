package models

import "github.com/google/uuid"

type ApologyAttachment struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ApologyID uuid.UUID `gorm:"type:uuid;not null" json:"apology_id"`
	FileName  string    `json:"file_name"`
	FileURL   string    `json:"file_url"`
	PublicID  string    `json:"public_id"`
	Size      string    `json:"size"`
}

func (ApologyAttachment) TableName() string {
	return "apology_attachments"
}
