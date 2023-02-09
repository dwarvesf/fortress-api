package invoice

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get invoice by id
func (s *store) One(db *gorm.DB, id string) (*model.Invoice, error) {
	var invoice *model.Invoice
	return invoice, db.Where("id = ?", id).First(&invoice).Error
}

// IsExist check the existence of invoice
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM invoices WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

func (s *store) GetLatestInvoiceByProject(db *gorm.DB, projectID string) (*model.Invoice, error) {
	iv := model.Invoice{}
	return &iv, db.Where("project_id = ? AND status != ? AND status != ?", projectID, model.InvoiceDraft, model.InvoiceError).Order("created_at DESC").First(&iv).Error
}

// All get all invoice
func (s *store) All(db *gorm.DB) ([]*model.Invoice, error) {
	var invoice []*model.Invoice
	return invoice, db.Find(&invoice).Error
}

// Delete delete 1 invoice by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.Invoice{}).Error
}

// Create creates a new invoice
func (s *store) Create(db *gorm.DB, e *model.Invoice) (invoice *model.Invoice, err error) {
	return e, db.Create(e).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, invoice *model.Invoice) (*model.Invoice, error) {
	return invoice, db.Model(&invoice).Where("id = ?", invoice.ID).Updates(&invoice).First(&invoice).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Invoice, updatedFields ...string) (*model.Invoice, error) {
	invoice := model.Invoice{}
	return &invoice, db.Model(&invoice).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
