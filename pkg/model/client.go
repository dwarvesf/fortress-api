package model

// Client store information of a Client
type Client struct {
	BaseModel

	Name               string
	Description        string
	RegistrationNumber string
	Address            string
	Country            string
	Industry           string
	Website            string
	Contacts           []ClientContact
}
