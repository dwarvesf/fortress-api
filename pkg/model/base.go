package model

import (
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// BaseModel base model for domain type
type BaseModel struct {
	ID        UUID       `sql:",type:uuid" json:"id"`
	CreatedAt time.Time  `sql:"default:now()" json:"created_at"`
	UpdatedAt *time.Time `sql:"default:now()" json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// BeforeCreate prepare data before create data
func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	// tx.Scopes().SetSetColumn("ID", uuid.NewV4())
	tx.Model(m).Update("id", uuid.NewV4())
	tx.Model(m).Update("created_at", time.Now())
	return nil
}

// BeforeCreate prepare data before create data
func (m *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	tx.Model(m).Update("updated_at", time.Now())
	return nil
}
