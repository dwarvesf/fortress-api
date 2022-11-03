package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type PaginationResponse struct {
	model.Pagination
	Total int64 `json:"total"`
}
