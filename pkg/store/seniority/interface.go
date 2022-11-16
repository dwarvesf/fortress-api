package seniority

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) ([]*model.Seniority, error)
	One(db *gorm.DB, id model.UUID) (seniorities *model.Seniority, err error)
	Exists(db *gorm.DB, id string) (bool, error)
}
