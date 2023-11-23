package model

const (
	RoleFullTimeCode    = "full-time"
	RoleProjectLeadCode = "project-lead"
)

type Role struct {
	BaseModel

	Name   string `json:"name"`
	Code   string `json:"code"`
	Level  int64  `json:"level"`
	Color  string `json:"color"`
	IsShow bool   `json:"isShow"`

	Employees []Employee `gorm:"many2many:employee_roles;"`
}
