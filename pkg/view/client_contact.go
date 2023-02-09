package view

import (
	"encoding/json"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type ClientContact struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Role          string   `json:"role"`
	Emails        []string `json:"emails"`
	IsMainContact bool     `json:"isMainContact"`
}

func toClientContact(cc *model.ClientContact) *ClientContact {
	emails := []string{}
	if err := json.Unmarshal(cc.Emails, &emails); err != nil {
		logger.L.Error(err, "failed to scan emails")
	}

	return &ClientContact{
		ID:            cc.ID.String(),
		Name:          cc.Name,
		Role:          cc.Role,
		Emails:        emails,
		IsMainContact: cc.IsMainContact,
	}
}
