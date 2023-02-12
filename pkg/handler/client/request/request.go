package request

type CreateClientInput struct {
	Name               string                      `json:"name"`
	Description        string                      `json:"description"`
	RegistrationNumber string                      `json:"registrationNumber"`
	Address            string                      `json:"address"`
	Country            string                      `json:"country"`
	Industry           string                      `json:"industry"`
	Website            string                      `json:"website"`
	Contacts           []*CreateClientContactInput `json:"contacts"`
}

type CreateClientContactInput struct {
	Name          string   `json:"name"`
	Role          string   `json:"role"`
	Emails        []string `json:"emails"`
	IsMainContact bool     `json:"isMainContact"`
}

type UpdateClientInput struct {
	Name               string                      `json:"name"`
	Description        string                      `json:"description"`
	RegistrationNumber string                      `json:"registrationNumber"`
	Address            string                      `json:"address"`
	Country            string                      `json:"country"`
	Industry           string                      `json:"industry"`
	Website            string                      `json:"website"`
	Contacts           []*UpdateClientContactInput `json:"contacts"`
}

type UpdateClientContactInput struct {
	Name          string   `json:"name"`
	Role          string   `json:"role"`
	Emails        []string `json:"emails"`
	IsMainContact bool     `json:"isMainContact"`
}
