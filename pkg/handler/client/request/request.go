package request

type CreateClientRequest struct {
	Name               string                      `json:"name"`
	Description        string                      `json:"description"`
	RegistrationNumber string                      `json:"registrationNumber"`
	Address            string                      `json:"address"`
	Country            string                      `json:"country"`
	Industry           string                      `json:"industry"`
	Website            string                      `json:"website"`
	Contacts           []*CreateClientContactInput `json:"contacts"`
} // @name CreateClientRequest

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
} // @name UpdateClientInput

type UpdateClientContactInput struct {
	Name          string   `json:"name"`
	Role          string   `json:"role"`
	Emails        []string `json:"emails"`
	IsMainContact bool     `json:"isMainContact"`
}
