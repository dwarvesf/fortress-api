package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type Content struct {
	ID        UUID       `sql:",type:uuid" json:"id"`
	CreatedAt time.Time  `sql:"default:now()" json:"createdAt"`
	UpdatedAt *time.Time `sql:"default:now()" json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`

	Type       string `json:"type"`
	Extension  string `json:"extension"`
	Path       string `json:"path"`
	UploadBy   UUID   `json:"uploadBy"`
	EmployeeID UUID   `json:"employee_id"`
}

// BeforeCreate prepare data before create data
func (m *Content) BeforeCreate(tx *gorm.DB) error {
	m.ID = UUID(uuid.NewV4())
	return nil
}

type ContentExtension string

const (
	ContentExtensionJpg ContentExtension = ".jpg"
	ContentExtensionPng ContentExtension = ".png"
	ContentExtensionPdf ContentExtension = ".pdf"
)

const (
	MaxFileSizeImage = 1000000
	MaxFileSizePdf   = 5000000
)

func (e ContentExtension) Valid() bool {
	switch e {
	case
		ContentExtensionJpg,
		ContentExtensionPng,
		ContentExtensionPdf:
		return true
	}
	return false
}
func (e ContentExtension) ImageValid() bool {
	switch e {
	case
		ContentExtensionJpg,
		ContentExtensionPng:
		return true
	}
	return false
}

func (e ContentExtension) String() string {
	return string(e)
}
