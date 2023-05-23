package model

type EmployeeInvitation struct {
	BaseModel

	EmployeeID               UUID
	InvitedBy                UUID
	InvitationCode           string
	IsCompleted              bool
	IsInfoUpdated            bool
	IsDiscordRoleAssigned    bool
	IsBasecampAccountCreated bool
	IsTeamEmailCreated       bool
	Employee                 *Employee
}

type InvitationEmail struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Link    string `json:"link"`
	Inviter string `json:"inviter"`
}
