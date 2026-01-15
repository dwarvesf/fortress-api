package view

import (
	"encoding/json"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type CreateInvoiceResponse struct {
	Data *model.Invoice `json:"data"`
}

type Invoice struct {
	ID               string        `json:"id"`
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
	SubTotal         float64       `json:"subTotal"`
	Tax              float64       `json:"tax"`
	Discount         float64       `json:"discount"`
	Total            float64       `json:"total"`
	ConversionAmount float64       `json:"conversionAmount"`
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
} // @name Invoice

type ClientInfo struct {
	ClientCompany string              `json:"clientCompany"`
	ClientAddress string              `json:"clientAddress"`
	Contacts      []ClientContactInfo `json:"contacts"`
} // @name ClientInfo

type ClientContactInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Role          string   `json:"-"`
	Emails        []string `json:"emails"`
	IsMainContact bool     `json:"isMainContact"`
} // @name ClientContactInfo

type CompanyInfo struct {
	ID                 string                        `json:"id"`
	Name               string                        `json:"name"`
	Description        string                        `json:"description"`
	RegistrationNumber string                        `json:"registrationNumber"`
	Info               map[string]CompanyContactInfo `json:"info"`
} // @name CompanyInfo

type CompanyContactInfo struct {
	Address string `json:"address"`
	Phone   string `json:"phone"`
} // @name CompanyContactInfo

func ToCompanyInfo(c *model.CompanyInfo) *CompanyInfo {
	if c == nil {
		return nil
	}

	companyContact := make(map[string]CompanyContactInfo)
	err := json.Unmarshal(c.Info.Bytes, &companyContact)
	if err != nil {
		return nil
	}

	return &CompanyInfo{
		ID:                 c.ID.String(),
		Name:               c.Name,
		Description:        c.Description,
		RegistrationNumber: c.RegistrationNumber,
		Info:               companyContact,
	}
}

type InvoiceItem struct {
	Quantity    float64 `json:"quantity"`
	UnitCost    float64 `json:"unitCost"`
	Discount    float64 `json:"discount"`
	Cost        float64 `json:"cost"`
	Description string  `json:"description"`
	IsExternal  bool    `json:"isExternal"`
} // @name InvoiceItem

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
} // @name ProjectInvoiceTemplate

type InvoiceTemplateResponse struct {
	Data ProjectInvoiceTemplate `json:"data"`
} // @name InvoiceTemplateResponse

func ToInvoiceInfo(invoice *model.Invoice) (*Invoice, error) {
	if invoice != nil {
		cc := make([]string, 0)
		if invoice.CC != nil {
			if err := json.Unmarshal(invoice.CC, &cc); err != nil {
				logger.L.Errorf(err, "failed to unmarshal CC for invoice %s", invoice.Number)
			}
		}

		invoiceItems, err := toInvoiceItem(invoice.LineItems)
		if err != nil {
			logger.L.Errorf(err, "failed to unmarshal line items for invoice %s", invoice.Number)
			invoiceItems = []InvoiceItem{}
		}

		rs := &Invoice{
			ID:               invoice.ID.String(),
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
		companyContact := make(map[string]CompanyContactInfo)
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

type InvoiceData struct {
	Invoice
	ProjectName string      `json:"projectName"`
	BankAccount BankAccount `json:"bankAccount"`
	CompanyInfo CompanyInfo `json:"companyInfo"`
	Client      ClientInfo  `json:"client"`
} // @name InvoiceData

type InvoiceListResponse struct {
	Total int64         `json:"total"`
	Page  int64         `json:"page"`
	Size  int64         `json:"size"`
	Sort  string        `json:"sort"`
	Data  []InvoiceData `json:"data"`
} // @name InvoiceListResponse

func ToInvoiceListResponse(invoices []*model.Invoice) ([]InvoiceData, error) {
	rs := make([]InvoiceData, 0)

	for _, invoice := range invoices {
		iv, err := ToInvoiceInfo(invoice)
		if err != nil {
			logger.L.Errorf(err, "failed to convert invoice %s", invoice.Number)
			continue
		}

		bankAccount := BankAccount{}
		if invoice.Bank != nil {
			currency := Currency{}
			if invoice.Bank.Currency != nil {
				currency = Currency{
					ID:     invoice.Bank.Currency.ID.String(),
					Name:   invoice.Bank.Currency.Name,
					Symbol: invoice.Bank.Currency.Symbol,
					Locale: invoice.Bank.Currency.Locale,
					Type:   invoice.Bank.Currency.Type,
				}
			}
			bankAccount = BankAccount{
				ID:            invoice.Bank.ID.String(),
				AccountNumber: invoice.Bank.AccountNumber,
				BankName:      invoice.Bank.BankName,
				OwnerName:     invoice.Bank.OwnerName,
				Address:       invoice.Bank.Address,
				SwiftCode:     invoice.Bank.SwiftCode,
				RoutingNumber: invoice.Bank.RoutingNumber,
				Name:          invoice.Bank.Name,
				UKSortCode:    invoice.Bank.UKSortCode,
				CurrencyID:    invoice.Bank.CurrencyID.String(),
				Currency:      currency,
			}
		}

		companyInfo := CompanyInfo{}
		clientInfo := ClientInfo{}
		if invoice.Project != nil {
			if invoice.Project.CompanyInfo != nil {
				companyContact := make(map[string]CompanyContactInfo)
				if err := json.Unmarshal(invoice.Project.CompanyInfo.Info.Bytes, &companyContact); err != nil {
					logger.L.Errorf(err, "failed to unmarshal company info for invoice %s", invoice.Number)
				}

				companyInfo = CompanyInfo{
					ID:                 invoice.Project.CompanyInfo.ID.String(),
					Name:               invoice.Project.CompanyInfo.Name,
					Description:        invoice.Project.CompanyInfo.Description,
					RegistrationNumber: invoice.Project.CompanyInfo.RegistrationNumber,
					Info:               companyContact,
				}
			}

			if invoice.Project.Client != nil {
				contacts := make([]ClientContactInfo, 0)
				for _, c := range invoice.Project.Client.Contacts {
					emails := make([]string, 0)
					if err := json.Unmarshal(c.Emails, &emails); err != nil {
						logger.L.Errorf(err, "failed to unmarshal emails for contact %s in invoice %s", c.Name, invoice.Number)
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
					ClientCompany: invoice.Project.Client.Name,
					ClientAddress: invoice.Project.Client.Address,
					Contacts:      contacts,
				}
			}
		}

		projectName := ""
		if invoice.Project != nil {
			projectName = invoice.Project.Name
		}

		rs = append(rs, InvoiceData{
			Invoice:     *iv,
			ProjectName: projectName,
			BankAccount: bankAccount,
			CompanyInfo: companyInfo,
			Client:      clientInfo,
		})
	}

	return rs, nil
}

// GenerateSplitsResponse is the response for generating invoice splits
type GenerateSplitsResponse struct {
	LegacyNumber  string `json:"legacy_number"`
	InvoicePageID string `json:"invoice_page_id"`
	JobEnqueued   bool   `json:"job_enqueued"`
	Message       string `json:"message"`
} // @name GenerateSplitsResponse
