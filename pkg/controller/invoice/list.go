package invoice

import (
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store/invoice"
)

type GetListInvoiceInput struct {
	model.Pagination
	ProjectIDs    []string
	Statuses      []string
	InvoiceNumber string
}

func (c *controller) List(in GetListInvoiceInput) ([]*model.Invoice, int64, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "invoice",
		"method":     "List",
		"input":      in,
	})

	invoices, total, err := c.store.Invoice.All(c.repo.DB(), invoice.GetInvoicesFilter{
		Preload:       true,
		ProjectIDs:    in.ProjectIDs,
		Statuses:      in.Statuses,
		InvoiceNumber: in.InvoiceNumber,
	}, in.Pagination)
	if err != nil {
		l.Error(err, "failed to get invoice list")
		return nil, 0, err
	}

	return invoices, total, nil
}
