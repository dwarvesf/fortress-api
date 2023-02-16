package model

import (
	"gorm.io/datatypes"
)

// CompanyInfo contain company information
type CompanyInfo struct {
	BaseModel

	Name               string         `json:"name"`
	Description        string         `json:"description"`
	RegistrationNumber string         `json:"registration_number"`
	Info               datatypes.JSON `json:"info"`
}

type CompanyContactInfo struct {
	Address string `json:"address"`
	Phone   string `json:"phone"`
}
