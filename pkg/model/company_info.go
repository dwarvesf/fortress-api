package model

import "github.com/jackc/pgtype"

// CompanyInfo contain company information
type CompanyInfo struct {
	BaseModel

	Name               string       `json:"name"`
	Description        string       `json:"description"`
	RegistrationNumber string       `json:"registrationNumber"`
	Info               pgtype.JSONB `json:"info"`
}

type CompanyContactInfo struct {
	Address string `json:"address"`
	Phone   string `json:"phone"`
}
