package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type Currency struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Locale string `json:"locale"`
	Type   string `json:"type"`
}

func toCurrency(c *model.Currency) Currency {
	if c == nil {
		return Currency{}
	}
	return Currency{
		ID:     c.ID.String(),
		Name:   c.Name,
		Symbol: c.Symbol,
		Locale: c.Locale,
		Type:   c.Type,
	}
}
