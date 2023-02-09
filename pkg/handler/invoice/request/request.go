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

type UpdateInput struct {
	Status model.InvoiceStatus `json:"status"`
}

func (r *UpdateInput) Validate() error {
	if r.Status != "" && !r.Status.IsValid() {
		return errs.ErrInvalidInvoiceStatus
	}

	return nil
}

type GetLatestInvoiceInput struct {
	ProjectID string `json:"projectID" form:"projectID"`
}
