package request

import (
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils"
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

type GetListInvoiceInput struct {
	model.Pagination
	ProjectID []string `json:"projectID" form:"projectID"`
	Status    []string `json:"status" form:"status"`
}

func (r *GetListInvoiceInput) StandardizeInput() {
	statuses := utils.RemoveEmptyString(r.Status)
	projectsIDs := utils.RemoveEmptyString(r.ProjectID)
	r.Pagination.Standardize()
	r.Status = statuses
	r.ProjectID = projectsIDs
}

func (r *GetListInvoiceInput) Validate() error {
	for _, status := range r.Status {
		if !model.InvoiceStatus(status).IsValid() {
			return errs.ErrInvalidInvoiceStatus
		}
	}

	for _, ids := range r.ProjectID {
		if _, err := model.UUIDFromString(ids); err != nil {
			return errs.ErrInvalidProjectID
		}
	}

	return nil
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
	Total       float64       `json:"total" binding:"gte=0"`
	Discount    float64       `json:"discount" binding:"gte=0"`
	Tax         float64       `json:"tax" binding:"gte=0"`
	SubTotal    float64       `json:"subtotal" binding:"gte=0"`
	InvoiceDate string        `json:"invoiceDate" binding:"required"`
	DueDate     string        `json:"dueDate" binding:"required"`
	Month       int           `json:"invoiceMonth" binding:"gte=0,lte=11"`
	Year        int           `json:"invoiceYear" binding:"gte=0"`
	SentByID    *model.UUID
	Number      string
}

type InvoiceItem struct {
	Quantity    float64 `json:"quantity"`
	UnitCost    float64 `json:"unitCost"`
	Discount    float64 `json:"discount"`
	Cost        float64 `json:"cost"`
	Description string  `json:"description"`
	IsExternal  bool    `json:"isExternal"`
}

func toInvoiceItemsModel(lineItems []InvoiceItem) []model.InvoiceItem {
	var items []model.InvoiceItem
	for _, item := range lineItems {
		items = append(items, model.InvoiceItem{
			Quantity:    math.Round(item.Quantity*100) / 100,
			UnitCost:    math.Round(item.UnitCost*100) / 100,
			Discount:    math.Round(item.Discount*100) / 100,
			Cost:        math.Round(item.Cost*100) / 100,
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

	dueAt, err := time.Parse("2006-01-02", i.DueDate)
	if err != nil {
		return nil, err
	}

	invoiceAt, err := time.Parse("2006-01-02", i.InvoiceDate)
	if err != nil {
		return nil, err
	}

	defaultStatus := model.InvoiceStatusSent
	if i.IsDraft {
		defaultStatus = model.InvoiceStatusDraft
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
		Total:       math.Round(i.Total*100) / 100,
		Discount:    math.Round(i.Discount*100) / 100,
		Tax:         i.Tax,
		SubTotal:    math.Round(i.SubTotal*100) / 100,
		Month:       i.Month + 1,
		Year:        i.Year,
		Status:      defaultStatus,
		SentBy:      i.SentByID,
		DueAt:       &dueAt,
		InvoicedAt:  &invoiceAt,
	}, nil
}
