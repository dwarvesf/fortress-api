package view

type PaginationResponse struct {
	Pagination
	Total int64 `json:"total"`
} // @name PaginationResponse
