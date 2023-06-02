package icydistribution

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// IStore is an interface for icydistribution store
type IStore interface {
	GetWeekly(db *gorm.DB) ([]model.IcyDistribution, error)
}
