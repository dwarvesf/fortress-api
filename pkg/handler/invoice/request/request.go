package request

import (
	"regexp"

	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type CreateInput struct {
	Number           string              `json:"number"`
	DueAt            string              `json:"dueAt"`
	PaidAt           string              `json:"paidAt"`
	Status           model.InvoiceStatus `json:"status"`
	Email            string              `json:"email"`
	Description      string              `json:"description"`
	Note             string              `json:"note"`
	SubTotal         int64               `json:"subTotal"`
	Tax              int64               `json:"tax"`
	Discount         int64               `json:"discount"`
	Total            int64               `json:"total"`
	ConversionAmount int64               `json:"conversionAmount"`
	Month            int                 `json:"month"`
	Year             int                 `json:"year"`
	ThreadID         string              `json:"threadID"`
	ScheduledDate    string              `json:"scheduledDate"`
	ConversionRate   float64             `json:"conversionRate"`
	BankID           model.UUID          `json:"bankID" binding:"required"`
	ProjectID        model.UUID          `json:"projectID"`
}

func (r *CreateInput) Validate() error {
	if r.Status != "" && !r.Status.IsValid() {
		return errs.ErrInvalidInvoiceStatus
	}

	regex, _ := regexp.Compile(".+@.+\\..+")
	if r.Email != "" && !regex.MatchString(r.Email) {
		return errs.ErrInvalidEmailDomain
	}

	return nil
}

type UpdateStatusRequest struct {
	Status            model.InvoiceStatus `json:"status"`
	SendThankYouEmail bool                `json:"sendThankYouEmail"`
}

func (r *UpdateStatusRequest) Validate() error {
	if r.Status != "" && !r.Status.IsValid() {
		return errs.ErrInvalidInvoiceStatus
	}

	return nil
}

type GetInvoiceInput struct {
	ProjectID string `json:"projectID" form:"projectID"`
}

type SendInvoiceRequest struct {
	IsDraft     bool       `json:"isDraft"`
	ProjectID   model.UUID `json:"projectID" binding:"required"`
	BankID      model.UUID `json:"bankID" binding:"required"`
	Description string     `json:"description"`
	Note        string     `json:"note"`
	CC          model.JSON `json:"cc"`
	LineItems   model.JSON `json:"lineItems"`
	Email       string     `json:"email" binding:"required,email"`
	Total       int        `json:"total" binding:"gte=0"`
	Discount    int        `json:"discount" binding:"gte=0"`
	Tax         int        `json:"tax" binding:"gte=0"`
	SubTotal    int        `json:"subtotal" binding:"gte=0"`
	InvoiceDate string     `json:"invoiceDate" binding:"required"`
	DueDate     string     `json:"dueDate" binding:"required"`
	Month       int        `json:"invoiceMonth" binding:"gte=0,lte=11"`
	Year        int        `json:"invoiceYear" binding:"gte=0"`
	SentByID    *model.UUID
	Number      string
}

func (i *SendInvoiceRequest) ToInvoiceModel() (*model.Invoice, error) {
	return &model.Invoice{
		ProjectID:   i.ProjectID,
		BankID:      i.BankID,
		Description: i.Description,
		Note:        i.Note,
		LineItems:   i.LineItems,
		Email:       i.Email,
		CC:          i.CC,
		Total:       int64(i.Total),
		Discount:    int64(i.Discount),
		Tax:         int64(i.Tax),
		SubTotal:    int64(i.SubTotal),
		Month:       i.Month + 1,
		Year:        i.Year,
		Status:      model.InvoiceStatusSent,
		SentBy:      i.SentByID,
	}, nil
}
