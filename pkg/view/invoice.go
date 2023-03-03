package view

import (
	"encoding/json"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type CreateInvoiceResponse struct {
	Data *model.Invoice `json:"data"`
}

type GetLatestInvoiceResponse struct {
	Data *model.Invoice `json:"data"`
}

type Invoice struct {
	Number           string        `json:"number"`
	InvoicedAt       *time.Time    `json:"invoicedAt"`
	DueAt            *time.Time    `json:"dueAt"`
	PaidAt           *time.Time    `json:"paidAt"`
	FailedAt         *time.Time    `json:"failedAt"`
	Status           string        `json:"status"`
	Email            string        `json:"email"`
	CC               []string      `json:"cc"`
	Description      string        `json:"description"`
	Note             string        `json:"note"`
	SubTotal         int64         `json:"subTotal"`
	Tax              int64         `json:"tax"`
	Discount         int64         `json:"discount"`
	Total            int64         `json:"total"`
	ConversionAmount int64         `json:"conversionAmount"`
	InvoiceFileURL   string        `json:"invoiceFileURL"`
	ErrorInvoiceID   string        `json:"errorInvoiceID"`
	LineItems        []InvoiceItem `json:"lineItems"`
	Month            int           `json:"month"`
	Year             int           `json:"year"`
	SentBy           string        `json:"sentBy"`
	ThreadID         string        `json:"threadID"`
	ScheduledDate    *time.Time    `json:"scheduledDate"`
	ConversionRate   float64       `json:"conversionRate"`
	BankID           string        `json:"bankID"`
	ProjectID        string        `json:"projectID"`
}

type ClientInfo struct {
	ClientCompany string              `json:"clientCompany"`
	ClientAddress string              `json:"clientAddress"`
	Contacts      []ClientContactInfo `json:"contacts"`
}

type ClientContactInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Role          string   `json:"-"`
	Emails        []string `json:"emails"`
	IsMainContact bool     `json:"isMainContact"`
}

type CompanyInfo struct {
	ID                 string                              `json:"id"`
	Name               string                              `json:"name"`
	Description        string                              `json:"description"`
	RegistrationNumber string                              `json:"registrationNumber"`
	Info               map[string]model.CompanyContactInfo `json:"info"`
}

type InvoiceItem struct {
	Quantity    float64 `json:"quantity"`
	UnitCost    int64   `json:"unitCost"`
	Discount    int64   `json:"discount"`
	Cost        int64   `json:"cost"`
	Description string  `json:"description"`
	IsExternal  bool    `json:"isExternal"`
}

func toInvoiceItem(lineItems model.JSON) ([]InvoiceItem, error) {
	var items []InvoiceItem
	var tmp []model.InvoiceItem

	if len(lineItems) == 0 || string(lineItems) == "null" {
		return items, nil
	}

	if err := json.Unmarshal(lineItems, &tmp); err != nil {
		return nil, err
	}

	for _, item := range tmp {
		items = append(items, InvoiceItem{
			Quantity:    item.Quantity,
			UnitCost:    item.UnitCost,
			Discount:    item.Discount,
			Cost:        item.Cost,
			Description: item.Description,
			IsExternal:  item.IsExternal,
		})
	}

	return items, nil
}

type ProjectInvoiceTemplate struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	InvoiceNumber string      `json:"invoiceNumber"`
	LastInvoice   *Invoice    `json:"lastInvoice"`
	Client        ClientInfo  `json:"client"`
	BankAccount   BankAccount `json:"bankAccount"`
	CompanyInfo   CompanyInfo `json:"companyInfo"`
}

type InvoiceTemplateResponse struct {
	Data ProjectInvoiceTemplate `json:"data"`
}

func ToInvoiceInfo(invoice *model.Invoice) (*Invoice, error) {
	if invoice != nil {

		cc := make([]string, 0)
		err := json.Unmarshal(invoice.CC, &cc)
		if err != nil {
			return nil, err
		}

		invoiceItems, err := toInvoiceItem(invoice.LineItems)
		if err != nil {
			return nil, err
		}

		rs := &Invoice{
			Number:           invoice.Number,
			InvoicedAt:       invoice.InvoicedAt,
			DueAt:            invoice.DueAt,
			PaidAt:           invoice.PaidAt,
			FailedAt:         invoice.FailedAt,
			Status:           invoice.Status.String(),
			Email:            invoice.Email,
			CC:               cc,
			Description:      invoice.Description,
			Note:             invoice.Note,
			SubTotal:         invoice.SubTotal,
			Tax:              invoice.Tax,
			Discount:         invoice.Discount,
			Total:            invoice.Total,
			ConversionAmount: invoice.ConversionAmount,
			InvoiceFileURL:   invoice.InvoiceFileURL,
			LineItems:        invoiceItems,
			Month:            invoice.Month,
			Year:             invoice.Year,
			ThreadID:         invoice.ThreadID,
			ScheduledDate:    invoice.ScheduledDate,
			ConversionRate:   invoice.ConversionRate,
			BankID:           invoice.BankID.String(),
			ProjectID:        invoice.ProjectID.String(),
		}

		if invoice.SentBy != nil {
			rs.SentBy = invoice.SentBy.String()
		}

		if invoice.ErrorInvoiceID != nil {
			rs.ErrorInvoiceID = invoice.ErrorInvoiceID.String()
		}

		return rs, nil
	}

	return nil, nil
}
func ToInvoiceTemplateResponse(p *model.Project, lastInvoice *model.Invoice, nextInvoiceNUmber string) (*ProjectInvoiceTemplate, error) {

	companyInfo := CompanyInfo{}
	if p.CompanyInfo != nil {
		companyContact := make(map[string]model.CompanyContactInfo)
		_ = json.Unmarshal(p.CompanyInfo.Info.Bytes, &companyContact)

		companyInfo = CompanyInfo{
			ID:                 p.CompanyInfo.ID.String(),
			Name:               p.CompanyInfo.Name,
			Description:        p.CompanyInfo.Description,
			RegistrationNumber: p.CompanyInfo.RegistrationNumber,
			Info:               companyContact,
		}
	}

	clientInfo := ClientInfo{}
	if p.Client != nil {

		contacts := make([]ClientContactInfo, 0)
		for _, c := range p.Client.Contacts {

			emails := make([]string, 0)
			err := json.Unmarshal(c.Emails, &emails)
			if err != nil {
				return nil, err
			}

			contacts = append(contacts, ClientContactInfo{
				ID:            c.ID.String(),
				Name:          c.Name,
				Role:          c.Role,
				Emails:        emails,
				IsMainContact: c.IsMainContact,
			})
		}
		clientInfo = ClientInfo{
			ClientCompany: p.Client.Name,
			ClientAddress: p.Client.Address,
			Contacts:      contacts,
		}
	}

	bankAccount := BankAccount{}
	if p.BankAccount != nil {

		currency := Currency{
			ID:     p.BankAccount.Currency.ID.String(),
			Name:   p.BankAccount.Currency.Name,
			Symbol: p.BankAccount.Currency.Symbol,
			Locale: p.BankAccount.Currency.Locale,
			Type:   p.BankAccount.Currency.Type,
		}
		bankAccount = BankAccount{
			ID:            p.BankAccount.ID.String(),
			AccountNumber: p.BankAccount.AccountNumber,
			BankName:      p.BankAccount.BankName,
			OwnerName:     p.BankAccount.OwnerName,
			Address:       p.BankAccount.Address,
			SwiftCode:     p.BankAccount.SwiftCode,
			RoutingNumber: p.BankAccount.RoutingNumber,
			Name:          p.BankAccount.Name,
			UKSortCode:    p.BankAccount.UKSortCode,
			CurrencyID:    p.BankAccount.CurrencyID.String(),
			Currency:      currency,
		}
	}

	iv, err := ToInvoiceInfo(lastInvoice)
	if err != nil {
		return nil, err
	}
	return &ProjectInvoiceTemplate{
		ID:            p.ID.String(),
		Name:          p.Name,
		InvoiceNumber: nextInvoiceNUmber,
		LastInvoice:   iv,
		Client:        clientInfo,
		BankAccount:   bankAccount,
		CompanyInfo:   companyInfo,
	}, nil
}
