package dashboard

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	GetResourceUtilizationByYear(db *gorm.DB) ([]*model.ResourceUtilization, error)
}
