package model

// Client store information of a Client
type Client struct {
	BaseModel

	Name               string
	Description        string
	RegistrationNumber string
	Avatar             string
	Address            string
	Country            string
	City               string
	Industry           string
	Website            string
	IsPublic           bool
	Lat                string
	Long               string
	CompanySize        string
	SolutionType       string

	Contacts []ClientContact
	Projects []Project
}
