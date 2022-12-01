package payroll

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type PayrollInput struct {
	ID          model.UUID
	EmployeeID  model.UUID
	EmployeeIDs []model.UUID
	Year        int
	Month       time.Month
	Day         int
	IsNotCommit bool
	GetLatest   bool
}

type PayrollDashboardInput struct {
	Paydays     []string
	Contracts   []string
	Date        string
	Departments []string
}
