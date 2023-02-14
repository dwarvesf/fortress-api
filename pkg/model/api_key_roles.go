package model

type APIKeyRole struct {
	BaseModel

	APIKeyID UUID `gorm:"column:api_key_id;default:null"`
	RoleID   UUID

	ApiKey APIKey
	Role   Role
}
