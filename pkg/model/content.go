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
	TargetID   UUID   `json:"target_id"`
	TargetType string `json:"target_type"`
	AuthType   string `json:"auth_type"`
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

type ContentTargetType string

const (
	ContentTargetTypeEmployee  ContentTargetType = "employees"
	ContentTargetTypeProject   ContentTargetType = "projects"
	ContentTargetTypeChangeLog ContentTargetType = "change-logs"
	ContentTargetTypeInvoice   ContentTargetType = "invoices"
)

type ContentType string

const (
	ContentTypeImage ContentType = "image"
	ContentTypeDoc   ContentType = "doc"
)

const (
	MaxFileSizeImage = 2202099
	MaxFileSizePdf   = 5347737
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

func (e ContentTargetType) Valid() bool {
	switch e {
	case
		ContentTargetTypeEmployee,
		ContentTargetTypeProject,
		ContentTargetTypeInvoice,
		ContentTargetTypeChangeLog:
		return true
	}
	return false
}

func (e ContentTargetType) String() string {
	return string(e)
}

func (e ContentType) Valid() bool {
	switch e {
	case
		ContentTypeImage,
		ContentTypeDoc:
		return true
	}
	return false
}

func (e ContentType) String() string {
	return string(e)
}

type DocumentType string

const (
	DocumentTypeAvatar       DocumentType = "avatar"
	DocumentTypeIDPhotoFront DocumentType = "id_photo_front"
	DocumentTypeIDPhotoBack  DocumentType = "id_photo_back"
)

func (e DocumentType) Valid() bool {
	switch e {
	case
		DocumentTypeAvatar,
		DocumentTypeIDPhotoFront,
		DocumentTypeIDPhotoBack:
		return true
	}
	return false
}

func (e DocumentType) String() string {
	return string(e)
}
