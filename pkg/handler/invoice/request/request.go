package request

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/mailutils"
)

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
	IsDraft     bool          `json:"isDraft"`
	ProjectID   model.UUID    `json:"projectID" binding:"required"`
	BankID      model.UUID    `json:"bankID" binding:"required"`
	Description string        `json:"description"`
	Note        string        `json:"note"`
	CC          []string      `json:"cc"`
	LineItems   []InvoiceItem `json:"lineItems"`
	Email       string        `json:"email" binding:"required,email"`
	Total       int           `json:"total" binding:"gte=0"`
	Discount    int           `json:"discount" binding:"gte=0"`
	Tax         int           `json:"tax" binding:"gte=0"`
	SubTotal    int           `json:"subtotal" binding:"gte=0"`
	InvoiceDate string        `json:"invoiceDate" binding:"required"`
	DueDate     string        `json:"dueDate" binding:"required"`
	Month       int           `json:"invoiceMonth" binding:"gte=0,lte=11"`
	Year        int           `json:"invoiceYear" binding:"gte=0"`
	SentByID    *model.UUID
	Number      string
}

type InvoiceItem struct {
	Quantity    float64 `json:"quantity"`
	UnitCost    int64   `json:"unitCost"`
	Discount    int64   `json:"discount"`
	Cost        int64   `json:"cost"`
	Description string  `json:"description"`
	IsExternal  bool    `json:"isExternal"`
}

func toInvoiceItemsModel(lineItems []InvoiceItem) []model.InvoiceItem {
	var items []model.InvoiceItem
	for _, item := range lineItems {
		items = append(items, model.InvoiceItem{
			Quantity:    item.Quantity,
			UnitCost:    item.UnitCost,
			Discount:    item.Discount,
			Cost:        item.Cost,
			Description: item.Description,
			IsExternal:  item.IsExternal,
		})
	}

	return items
}

func (i *SendInvoiceRequest) ValidateAndMappingRequest(c *gin.Context, cfg *config.Config) error {
	if err := c.ShouldBindJSON(&i); err != nil {
		return err
	}

	var ccList []string
	for _, cc := range i.CC {
		if strings.TrimSpace(cc) == "" {
			continue
		}
		ccList = append(ccList, cc)
	}

	i.CC = ccList

	if cfg.Env == "prod" {
		return nil
	}

	if !mailutils.IsDwarvesMail(i.Email) {
		return errs.ErrInvalidDeveloperEmail
	}

	for _, v := range i.CC {
		if !mailutils.IsDwarvesMail(v) {
			return errs.ErrInvalidDeveloperEmail
		}
	}

	return nil
}

func (i *SendInvoiceRequest) ToInvoiceModel() (*model.Invoice, error) {
	lineItems, err := json.Marshal(toInvoiceItemsModel(i.LineItems))
	if err != nil {
		return nil, err
	}

	cc, err := json.Marshal(i.CC)
	if err != nil {
		return nil, err
	}

	return &model.Invoice{
		ProjectID:   i.ProjectID,
		BankID:      i.BankID,
		Description: i.Description,
		Note:        i.Note,
		LineItems:   lineItems,
		Email:       i.Email,
		CC:          cc,
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
