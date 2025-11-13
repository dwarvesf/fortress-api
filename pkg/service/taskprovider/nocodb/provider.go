package nocodb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/model"
	nocodbsvc "github.com/dwarvesf/fortress-api/pkg/service/nocodb"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
)

// Provider implements taskprovider.InvoiceProvider on top of NocoDB.
type Provider struct {
	svc *nocodbsvc.Service
}

// New creates a NocoDB-backed invoice provider.
func New(svc *nocodbsvc.Service) *Provider {
	if svc == nil {
		return nil
	}
	return &Provider{svc: svc}
}

func (p *Provider) Type() taskprovider.ProviderType {
	return taskprovider.ProviderNocoDB
}

func (p *Provider) EnsureTask(ctx context.Context, input taskprovider.CreateInvoiceTaskInput) (*taskprovider.InvoiceTaskRef, error) {
	if p == nil || p.svc == nil {
		return nil, errors.New("nocodb provider is not configured")
	}
	if input.Invoice == nil {
		return nil, errors.New("missing invoice data")
	}
	payload := buildInvoicePayload(input.Invoice)
	id, err := p.svc.UpsertInvoiceRecord(ctx, input.Invoice.Number, payload)
	if err != nil {
		return nil, err
	}
	return &taskprovider.InvoiceTaskRef{
		Provider:   taskprovider.ProviderNocoDB,
		ExternalID: id,
	}, nil
}

func (p *Provider) UploadAttachment(ctx context.Context, ref *taskprovider.InvoiceTaskRef, input taskprovider.InvoiceAttachmentInput) (*taskprovider.InvoiceAttachmentRef, error) {
	url := input.URL
	if url == "" {
		url = fmt.Sprintf("https://storage.googleapis.com/%s", input.FileName)
	}
	markup := fmt.Sprintf("[Invoice PDF](%s)", url)
	return &taskprovider.InvoiceAttachmentRef{ExternalID: url, Markup: markup}, nil
}

func (p *Provider) PostComment(ctx context.Context, ref *taskprovider.InvoiceTaskRef, input taskprovider.InvoiceCommentInput) error {
	if p == nil || p.svc == nil {
		return errors.New("nocodb provider is not configured")
	}
	if ref == nil || ref.ExternalID == "" {
		return errors.New("missing invoice reference")
	}
	return p.svc.CreateInvoiceComment(ctx, ref.ExternalID, "system", input.Message, input.Type)
}

func (p *Provider) CompleteTask(ctx context.Context, ref *taskprovider.InvoiceTaskRef) error {
	if p == nil || p.svc == nil {
		return errors.New("nocodb provider is not configured")
	}
	if ref == nil || ref.ExternalID == "" {
		return errors.New("missing invoice reference")
	}
	return p.svc.UpdateInvoiceStatus(ctx, ref.ExternalID, string(model.InvoiceStatusPaid))
}

func buildInvoicePayload(iv *model.Invoice) map[string]interface{} {
	currency := ""
	if iv.Bank != nil && iv.Bank.Currency != nil {
		currency = iv.Bank.Currency.Name
	} else if iv.Project != nil && iv.Project.BankAccount != nil && iv.Project.BankAccount.Currency != nil {
		currency = iv.Project.BankAccount.Currency.Name
	}
	payload := map[string]interface{}{
		"invoice_number":      iv.Number,
		"month":               iv.Month,
		"year":                iv.Year,
		"status":              string(iv.Status),
		"amount":              iv.Total,
		"currency":            currency,
		"fortress_invoice_id": iv.ID.String(),
	}
	if strings.TrimSpace(iv.InvoiceFileURL) != "" {
		payload["attachment_url"] = iv.InvoiceFileURL
	}
	return payload
}
