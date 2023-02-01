package socialaccount

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, sa *model.SocialAccount) (*model.SocialAccount, error)
	Update(db *gorm.DB, sa *model.SocialAccount) (*model.SocialAccount, error)
	GetByEmployeeID(db *gorm.DB, employeeID string) ([]*model.SocialAccount, error)
}
