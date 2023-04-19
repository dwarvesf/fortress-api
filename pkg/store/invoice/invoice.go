package invoice

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One getNext invoice by id
func (s *store) One(db *gorm.DB, query *Query) (*model.Invoice, error) {
	var invoice *model.Invoice
	if query.ID != "" {
		db = db.Where("id = ?", query.ID)
	}
	if query.Number != "" {
		db = db.Where("number = ?", query.Number)
	}
	return invoice, db.
		Preload("Project").
		Preload("Project.Heads", "deleted_at IS NULL AND (end_date IS NULL OR end_date > now())").
		Preload("Project.Heads.Employee", "deleted_at IS NULL").
		Preload("Project.BankAccount", "deleted_at IS NULL").
		Preload("Project.BankAccount.Currency", "deleted_at IS NULL").
		Preload("Project.Organization", "deleted_at IS NULL").
		First(&invoice).Error
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
	return &iv, db.Where("project_id = ? AND status != ? AND status != ?", projectID, model.InvoiceStatusDraft, model.InvoiceStatusError).Order("created_at DESC").First(&iv).Error
}

// All getNext all invoice
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

func (s *store) Save(db *gorm.DB, invoice *model.Invoice) (*model.Invoice, error) {
	return invoice, db.Save(&invoice).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Invoice, updatedFields ...string) (*model.Invoice, error) {
	invoice := model.Invoice{}
	return &invoice, db.Model(&invoice).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

func (s *store) GetNextInvoiceNumber(db *gorm.DB, year int, projectCode string) (*string, error) {
	yearComInvoiceKey := fmt.Sprintf("%s_%d", model.InvoiceCachingKey.YearInvoiceNumberPrefix, year)
	nextCompIvn, err := s.getNext(db, yearComInvoiceKey)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s_%s_%d", model.InvoiceCachingKey.ProjectInvoiceNumberPrefix, strings.ToUpper(projectCode), time.Now().Year())
	nextProjectIvn, err := s.getNext(db, key)
	if err != nil {
		return nil, err
	}
	invoiceNumberText := fmt.Sprintf("%d%d-%s-%03d",
		year,
		nextCompIvn,
		strings.ToUpper(projectCode),
		nextProjectIvn)

	return &invoiceNumberText, err
}

func (s *store) getNext(db *gorm.DB, key string) (next int, err error) {
	var maxNumber model.InvoiceNumberCaching
	err = db.Where(&model.InvoiceNumberCaching{Key: key}).Find(&maxNumber).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return maxNumber.Number + 1, nil
}
