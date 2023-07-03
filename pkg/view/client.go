package view

import (
	"strconv"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type CreateClientResponse struct {
	Data *model.Client `json:"data"`
}

type GetListClientResponse struct {
	Data []*model.Client `json:"data"`
}

type GetDetailClientResponse struct {
	Data *model.Client `json:"data"`
}

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

type Address struct {
	Address string `json:"address"`
	Country string `json:"country"`
	City    string `json:"city"`
	Lat     string `json:"lat"`
	Long    string `json:"long"`
}

type PublicClient struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Address      Address  `json:"address"`
	Stack        []string `json:"stacks"`
	Industry     string   `json:"industry"`
	CompanySize  string   `json:"companySize"`
	SolutionType string   `json:"solutionType"`
}

type PublicClientListResponse struct {
	Data []PublicClient `json:"data"`
}

func ToPublicClientListResponse(clients []*model.Client) []PublicClient {
	rs := make([]PublicClient, 0, len(clients))
	for _, client := range clients {
		stacks := make([]string, 0)

		for _, p := range client.Projects {
			for _, s := range p.ProjectStacks {
				stacks = append(stacks, s.Stack.Name)
			}
		}

		clientAddress := client.Country
		if strings.TrimSpace(client.City) != "" {
			clientAddress = client.City + ", " + clientAddress
		}

		rs = append(rs, PublicClient{
			ID:   client.ID.String(),
			Name: client.Name,
			Address: Address{
				Address: clientAddress,
				City:    client.City,
				Country: client.Country,
				Lat:     client.Lat,
				Long:    client.Long,
			},
			Stack:        stacks,
			Industry:     client.Industry,
			CompanySize:  strconv.Itoa(client.CompanySize),
			SolutionType: client.SolutionType,
		})
	}

	return rs
}
