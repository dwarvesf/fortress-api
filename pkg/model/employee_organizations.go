package model

type EmployeeOrganization struct {
	BaseModel

	EmployeeID     UUID
	OrganizationID UUID

	Organization Organization
}
