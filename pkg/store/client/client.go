package client

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get client by id
func (s *store) One(db *gorm.DB, id string) (*model.Client, error) {
	var client *model.Client
	return client, db.Where("id = ?", id).Preload("Contacts", "deleted_at IS NULL").First(&client).Error
}

// IsExist check client existence
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM clients WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

// All get all client
func (s *store) All(db *gorm.DB, public bool, preload bool) ([]*model.Client, error) {
	var client []*model.Client

	query := db.Preload("Contacts", "deleted_at IS NULL")

	if preload {
		query = query.
			Preload("Projects").
			Preload("Projects.ProjectStacks").
			Preload("Projects.ProjectStacks.Stack")
	}

	if public {
		query = query.Where("is_public = ?", true)
	}

	return client, query.Find(&client).Error
}

// Delete delete 1 client by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.Client{}).Error
}

// Create creates a new client
func (s *store) Create(db *gorm.DB, e *model.Client) (client *model.Client, err error) {
	return e, db.Create(e).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, client *model.Client) (*model.Client, error) {
	return client, db.Model(&client).Where("id = ?", client.ID).Updates(&client).First(&client).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Client, updatedFields ...string) (*model.Client, error) {
	client := model.Client{}
	return &client, db.Model(&client).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
