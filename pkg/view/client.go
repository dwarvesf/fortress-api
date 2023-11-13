package view

import (
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type CreateClientResponse struct {
	Data *Client `json:"data"`
} // @name CreateClientResponse

type GetListClientResponse struct {
	Data []*Client `json:"data"`
} // @name GetListClientResponse

type GetDetailClientResponse struct {
	Data *Client `json:"data"`
} // @name GetDetailClientResponse

type Client struct {
	ID                 string     `json:"id"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          *time.Time `json:"updatedAt"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	RegistrationNumber string     `json:"registrationNumber"`
	Avatar             string     `json:"avatar"`
	Address            string     `json:"address"`
	Country            string     `json:"country"`
	City               string     `json:"city"`
	Industry           string     `json:"industry"`
	Website            string     `json:"website"`
	IsPublic           bool       `json:"isPublic"`
	Lat                string     `json:"lat"`
	Long               string     `json:"long"`
	CompanySize        string     `json:"companySize"`
	SolutionType       string     `json:"solutionType"`

	Contacts []ClientContact `json:"contacts"`
	Projects []Project       `json:"projects"`
} // @name Client

func ToClients(clients []*model.Client) []Client {
	rs := make([]Client, 0, len(clients))
	for _, client := range clients {
		c := ToClient(client)
		if c != nil {
			rs = append(rs, *c)
		}
	}

	return rs
}

func ToClient(client *model.Client) *Client {
	contacts := make([]ClientContact, 0)
	for _, contact := range client.Contacts {
		contacts = append(contacts, *toClientContact(&contact))
	}

	return &Client{
		ID:                 client.ID.String(),
		CreatedAt:          client.CreatedAt,
		UpdatedAt:          client.UpdatedAt,
		Name:               client.Name,
		Description:        client.Description,
		RegistrationNumber: client.RegistrationNumber,
		Avatar:             client.Avatar,
		Address:            client.Address,
		Country:            client.Country,
		City:               client.City,
		Industry:           client.Industry,
		Website:            client.Website,
		IsPublic:           client.IsPublic,
		Lat:                client.Lat,
		Long:               client.Long,
		CompanySize:        client.CompanySize,
		SolutionType:       client.SolutionType,
		Contacts:           contacts,
		Projects:           ToProjects(client.Projects),
	}
}

type Address struct {
	Address string `json:"address"`
	Country string `json:"country"`
	City    string `json:"city"`
	Lat     string `json:"lat"`
	Long    string `json:"long"`
} // @name Address

type PublicClient struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Avatar       string   `json:"avatar"`
	Address      Address  `json:"address"`
	Stack        []string `json:"stacks"`
	Industry     string   `json:"industry"`
	CompanySize  string   `json:"companySize"`
	SolutionType string   `json:"solutionType"`
} // @name PublicClient

type PublicClientListResponse struct {
	Data []PublicClient `json:"data"`
} // @name PublicClientListResponse

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
			ID:     client.ID.String(),
			Name:   client.Name,
			Avatar: client.Avatar,
			Address: Address{
				Address: clientAddress,
				City:    client.City,
				Country: client.Country,
				Lat:     client.Lat,
				Long:    client.Long,
			},
			Stack:        stacks,
			Industry:     client.Industry,
			CompanySize:  client.CompanySize,
			SolutionType: client.SolutionType,
		})
	}

	return rs
}
