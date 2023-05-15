package model

import "time"

type EmployeeInvitation struct {
	BaseModel

	EmployeeID UUID
	InvitedBy  UUID
	Code       string
	ExpiredAt  *time.Time
}

type InvitationEmail struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Link    string `json:"link"`
	Inviter string `json:"inviter"`
}
