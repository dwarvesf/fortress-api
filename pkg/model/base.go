package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel base model for domain type
type BaseModel struct {
	ID        UUID            `sql:",type:uuid" json:"id" gorm:"default:uuid()"`
	CreatedAt time.Time       `sql:"default:now()" json:"createdAt"`
	UpdatedAt *time.Time      `sql:"default:now()" json:"updatedAt"`
	DeletedAt *gorm.DeletedAt `json:"deletedAt,omitempty"`
}

// BeforeCreate prepare data before create data
// func (m *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
// 	m.ID = NewUUID()
// 	m.CreatedAt = time.Now()
// 	return
// }

// BeforeUpdate prepare data before create data
func (m *BaseModel) BeforeUpdate(tx *gorm.DB) (err error) {
	tx.Statement.SetColumn("updated_at", time.Now())
	return
}
