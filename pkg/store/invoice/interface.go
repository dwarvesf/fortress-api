package invoice

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (invoice *model.Invoice, err error)
	IsExist(db *gorm.DB, id string) (exists bool, err error)
	All(db *gorm.DB) (invoices []*model.Invoice, err error)
	GetLatestInvoiceByProject(db *gorm.DB, projectID string) (invoice *model.Invoice, err error)
	GetNextInvoiceNumber(db *gorm.DB, year int, projectCode string) (*string, error)
	Delete(db *gorm.DB, id string) (err error)
	Create(db *gorm.DB, e *model.Invoice) (invoice *model.Invoice, err error)
	Save(db *gorm.DB, e *model.Invoice) (invoice *model.Invoice, err error)
	Update(db *gorm.DB, invoice *model.Invoice) (a *model.Invoice, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, invoice model.Invoice, updatedFields ...string) (a *model.Invoice, err error)
}
