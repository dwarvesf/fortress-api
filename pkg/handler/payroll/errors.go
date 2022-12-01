package payroll

import (
	"errors"
)

var (
	ErrInvalidPayrollDate = errors.New("invalid payroll date, month, year")
	ErrPayrollNotFound = errors.New("payroll not found")
)
