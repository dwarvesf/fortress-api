package view

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type Client struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	RegistrationNumber string          `json:"registrationNumber"`
	Address            string          `json:"address"`
	Country            string          `json:"country"`
	Industry           string          `json:"industry"`
	Website            string          `json:"website"`
	Contacts           []ClientContact `json:"contacts"`
}

func toClient(client *model.Client) *Client {
	contacts := make([]ClientContact, 0)
	for _, contact := range client.Contacts {
		contacts = append(contacts, *toClientContact(&contact))
	}

	return &Client{
		ID:                 client.ID.String(),
		Name:               client.Name,
		Description:        client.Description,
		RegistrationNumber: client.RegistrationNumber,
		Address:            client.Address,
		Country:            client.Country,
		Industry:           client.Industry,
		Website:            client.Website,
		Contacts:           contacts,
	}
}
