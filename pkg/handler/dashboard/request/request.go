package request

import "github.com/dwarvesf/fortress-api/pkg/model"

type WorkUnitDistributionInput struct {
	SortRequired model.SortOrder `form:"sortRequired" json:"sortRequired"`
}
