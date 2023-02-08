package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type CreateInvoiceResponse struct {
	Data *model.Invoice `json:"data"`
}

type GetLatestInvoiceResponse struct {
	Data *model.Invoice `json:"data"`
}
