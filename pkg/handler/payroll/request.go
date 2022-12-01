package payroll

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetListPayrollInput struct {
	model.Pagination

	Next  string     `form:"next" json:"next"`
	Batch int        `form:"batch" json:"batch"`
	Month time.Month `form:"month" json:"month"`
	Year  int        `form:"year" json:"year"`
}

func (i *GetListPayrollInput) Validate() error {
	if i.Batch < 0 || i.Month < 0 || i.Year < 0 {
		return ErrInvalidPayrollDate
	}
	return nil
}
