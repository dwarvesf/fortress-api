package employee

import "github.com/dwarvesf/fortress-api/pkg/model"

type GetListEmployeeQuery struct {
	model.Pagination

	WorkingStatus string `json:"workingStatus" form:"workingStatus"`
}
