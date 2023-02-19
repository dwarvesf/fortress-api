package model

type InvoiceNumberCaching struct {
	BaseModel

	Key    string `json:"key"`
	Number int    `json:"number"`
}

func (InvoiceNumberCaching) TableName() string { return "invoice_number_caching" }

type InvoiceCachingKeyStr struct {
	YearInvoiceNumberPrefix     string
	ProjectInvoiceNumberPrefix  string
	ProjectTemplateNumberPrefix string
	TplNumberPrefix             string
}

// InvoiceCachingKey present current keys of max numbers
var InvoiceCachingKey = InvoiceCachingKeyStr{
	YearInvoiceNumberPrefix:     "year_invoice",
	ProjectInvoiceNumberPrefix:  "project_invoice",
	ProjectTemplateNumberPrefix: "Project_Template_Number",
	TplNumberPrefix:             "Tpl",
}
