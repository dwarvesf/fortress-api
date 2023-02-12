package clientcontact

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get clientContact by id
func (s *store) One(db *gorm.DB, id string) (*model.ClientContact, error) {
	var clientContact *model.ClientContact
	return clientContact, db.Where("id = ?", id).First(&clientContact).Error
}

// IsExist check client contact existence
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM client_contacts WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

// All get all clientContact
func (s *store) All(db *gorm.DB) ([]*model.ClientContact, error) {
	var clientContact []*model.ClientContact
	return clientContact, db.Find(&clientContact).Error
}

// Delete delete 1 clientContact by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.ClientContact{}).Error
}

// Delete delete 1 clientContact by id
func (s *store) DeleteByClientID(db *gorm.DB, clientID string) error {
	return db.Where("client_id = ?", clientID).Delete(&model.ClientContact{}).Error
}

// Create creates a new clientContact
func (s *store) Create(db *gorm.DB, e *model.ClientContact) (clientContact *model.ClientContact, err error) {
	return e, db.Create(e).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, clientContact *model.ClientContact) (*model.ClientContact, error) {
	return clientContact, db.Model(&clientContact).Where("id = ?", clientContact.ID).Updates(&clientContact).First(&clientContact).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ClientContact, updatedFields ...string) (*model.ClientContact, error) {
	clientContact := model.ClientContact{}
	return &clientContact, db.Model(&clientContact).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
