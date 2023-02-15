package model

import (
	"gorm.io/datatypes"
)

// ClientContact is the model for the client_contact table
type ClientContact struct {
	BaseModel

	Name          string
	ClientID      UUID
	Role          string
	Emails        datatypes.JSON
	IsMainContact bool
}

type ClientEmail struct {
	Emails []string `json:"emails"`
}
