package model

type APIKey struct {
	BaseModel

	ClientID  string
	SecretKey string
	Status    APIKeyStatus

	ApikeyRoles []APIKeyRole
	Roles       []Role `gorm:"many2many:api_key_roles;"`
}

type APIKeyStatus string

// values for working_status
const (
	ApikeyStatusValid   APIKeyStatus = "valid"
	ApikeyStatusInvalid APIKeyStatus = "invalid"
)

// IsValid validation for APIKeyStatus
func (e APIKeyStatus) IsValid() bool {
	switch e {
	case
		ApikeyStatusValid,
		ApikeyStatusInvalid:
		return true
	}
	return false
}

type TokenType string

const (
	TokenTypeJWT    TokenType = "JWT"
	TokenTypeAPIKey TokenType = "ApiKey"
)

func (t TokenType) String() string {
	return string(t)
}
