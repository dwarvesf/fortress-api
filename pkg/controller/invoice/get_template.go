package invoice

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetInvoiceInput struct {
	Now       *time.Time
	ProjectID string
}

func (c *controller) GetTemplate(in GetInvoiceInput) (string, *model.Invoice, *model.Project, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "invoice",
		"method":     "GetTemplate",
		"input":      in,
	})

	// check project existence
	p, err := c.store.Project.One(c.repo.DB(), in.ProjectID, true)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(ErrProjectNotFound, "project not found")
		return "", nil, nil, ErrProjectNotFound
	}

	if err != nil {
		return "", nil, nil, err
	}

	nextInvoiceNumber, err := c.store.Invoice.GetNextInvoiceNumber(c.repo.DB(), in.Now.Year(), p.Code)
	if err != nil {
		l.Error(err, "failed to get next invoice Number")
		return "", nil, nil, ErrCouldNotGetTheNextInvoiceNumber
	}

	lastInvoice, err := c.store.Invoice.GetLatestInvoiceByProject(c.repo.DB(), in.ProjectID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(err, "failed to get the latest invoice")
		return "", nil, nil, ErrCouldNotGetTheLatestInvoice
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		lastInvoice = nil
	}

	return *nextInvoiceNumber, lastInvoice, p, nil
}
