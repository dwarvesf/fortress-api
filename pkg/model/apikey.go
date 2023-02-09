package model

type Apikey struct {
	BaseModel

	ClientID  string
	SecretKey string
	Status    ApikeyStatus

	ApikeyRoles []ApikeyRole
	Roles       []Role `gorm:"many2many:apikey_roles;"`
}

type ApikeyStatus string

// values for working_status
const (
	ApikeyStatusValid   ApikeyStatus = "valid"
	ApikeyStatusInvalid ApikeyStatus = "invalid"
)

// IsValid validation for ApikeyStatus
func (e ApikeyStatus) IsValid() bool {
	switch e {
	case
		ApikeyStatusValid,
		ApikeyStatusInvalid:
		return true
	}
	return false
}
