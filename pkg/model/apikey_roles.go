package model

type ApikeyRole struct {
	BaseModel

	ApiKeyID UUID `gorm:"column:apikey_id;default:null"`
	RoleID   UUID

	ApiKey Apikey
	Role   Role
}
