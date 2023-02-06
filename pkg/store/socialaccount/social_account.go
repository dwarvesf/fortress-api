package socialaccount

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create create new SocialAccount
func (s *store) Create(db *gorm.DB, sa *model.SocialAccount) (*model.SocialAccount, error) {
	return sa, db.Create(sa).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, sa *model.SocialAccount) (*model.SocialAccount, error) {
	return sa, db.Model(&sa).Where("id = ?", sa.ID).Updates(&sa).Error
}

func (s *store) GetByEmployeeID(db *gorm.DB, employeeID string) ([]*model.SocialAccount, error) {
	var accounts []*model.SocialAccount
	return accounts, db.Where("employee_id = ?", employeeID).Find(&accounts).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.SocialAccount, updatedFields ...string) (*model.SocialAccount, error) {
	sa := model.SocialAccount{}
	return &sa, db.Model(&sa).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
