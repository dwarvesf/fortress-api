package socialaccount

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, sa *model.SocialAccount) (account *model.SocialAccount, err error)
	Update(db *gorm.DB, sa *model.SocialAccount) (account *model.SocialAccount, err error)
	GetByEmployeeID(db *gorm.DB, employeeID string) (accounts []*model.SocialAccount, err error)
	GetByType(db *gorm.DB, saType string) (accounts []*model.SocialAccount, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.SocialAccount, updatedFields ...string) (*model.SocialAccount, error)
}
