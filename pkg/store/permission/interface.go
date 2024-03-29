package permission

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetByEmployeeID(db *gorm.DB, employeeID string) (permissions []*model.Permission, err error)
	GetByApiKeyID(db *gorm.DB, apiKeyID string) ([]*model.Permission, error)
	HasPermission(db *gorm.DB, employeeID string, perm string) (bool, error)
}
