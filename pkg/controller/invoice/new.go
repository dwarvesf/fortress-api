package invoice

import (
	"context"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

type controller struct {
	store   *store.Store
	service *service.Service
	worker  *worker.Worker
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, worker *worker.Worker, logger logger.Logger, cfg *config.Config) IController {
	return &controller{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
		config:  cfg,
		worker:  worker,
	}
}

type IController interface {
	GetTemplate(in GetInvoiceInput) (nextInvoiceNumber string, lastInvoice *model.Invoice, p *model.Project, err error)
	List(in GetListInvoiceInput) ([]*model.Invoice, int64, error)
	MarkInvoiceAsError(invoice *model.Invoice) (*model.Invoice, error)
	MarkInvoiceAsPaid(invoice *model.Invoice, sendThankYouEmail bool) (*model.Invoice, error)
	MarkInvoiceAsPaidByBasecampWebhookMessage(invoice *model.Invoice, msg *model.BasecampWebhookMessage) (*model.Invoice, error)
	MarkInvoiceAsPaidWithTaskRef(invoice *model.Invoice, ref *taskprovider.InvoiceTaskRef, sendThankYouEmail bool) (*model.Invoice, error)
	MarkInvoiceAsPaidByNumber(invoiceNumber string) (*MarkPaidResult, error)
	GenerateInvoiceSplitsByLegacyNumber(legacyNumber string) (*view.GenerateSplitsResponse, error)
	Send(iv *model.Invoice) (*model.Invoice, error)
	UpdateStatus(in UpdateStatusInput) (*model.Invoice, error)

	// Calculate commissions for an invoice
	CalculateCommissionFromInvoice(db store.DBRepo, l logger.Logger, invoice *model.Invoice) ([]model.EmployeeCommission, error)
	RemoveInboundFundCommission(employeeCommissions []model.EmployeeCommission) []model.EmployeeCommission
	ProcessCommissions(invoiceID string, dryRun bool, l logger.Logger) ([]model.EmployeeCommission, error)

	// Test method for PDF generation
	GenerateInvoicePDFForTest(l logger.Logger, invoice *model.Invoice, items []model.InvoiceItem) error

	// Generate PDF for Notion webhook invoice generation
	GenerateInvoicePDFForNotion(l logger.Logger, invoice *model.Invoice, items []model.InvoiceItem) error

	// Contractor invoice generation
	GenerateContractorInvoice(ctx context.Context, discord, month string) (*ContractorInvoiceData, error)
	GenerateContractorInvoicePDF(l logger.Logger, data *ContractorInvoiceData) ([]byte, error)
}
