package invoice

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB, filter GetInvoicesFilter, pagination model.Pagination) (invoices []*model.Invoice, total int64, err error)
	Create(db *gorm.DB, e *model.Invoice) (invoice *model.Invoice, err error)
	Delete(db *gorm.DB, id string) (err error)
	GetLatestInvoiceByProject(db *gorm.DB, projectID string) (invoice *model.Invoice, err error)
	GetNextInvoiceNumber(db *gorm.DB, year int, projectCode string) (*string, error)
	IsExist(db *gorm.DB, id string) (exists bool, err error)
	One(db *gorm.DB, query *Query) (invoice *model.Invoice, err error)
	Save(db *gorm.DB, e *model.Invoice) (invoice *model.Invoice, err error)
	Update(db *gorm.DB, invoice *model.Invoice) (a *model.Invoice, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, invoice model.Invoice, updatedFields ...string) (a *model.Invoice, err error)
}

// Query present invoice query from user
type Query struct {
	ID          string
	Alias       string
	Month       int64
	Year        int64
	ProjectName string
	Statuses    []model.InvoiceStatus
	Number      string
}

type GetInvoicesFilter struct {
	Preload       bool
	ProjectIDs    []string
	Statuses      []string
	InvoiceNumber string
}
