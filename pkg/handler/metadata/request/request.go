package request

import (
	"github.com/dwarvesf/fortress-api/pkg/handler/metadata/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type UpdateStackBody struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Avatar string `json:"avatar"`
}

type UpdateStackInput struct {
	ID   string
	Body UpdateStackBody
}

func (i UpdateStackInput) Validate() error {
	if i.ID == "" || !model.IsUUIDFromString(i.ID) {
		return errs.ErrInvalidStackID
	}

	return nil
}

type CreateStackInput struct {
	Name   string `json:"name" binding:"required"`
	Code   string `json:"code" binding:"required"`
	Avatar string `json:"avatar"`
}

type UpdatePositionBody struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type UpdatePositionInput struct {
	ID   string
	Body UpdatePositionBody
}

func (i UpdatePositionInput) Validate() error {
	if i.ID == "" || !model.IsUUIDFromString(i.ID) {
		return errs.ErrInvalidPositionID
	}

	return nil
}

type CreatePositionInput struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

type GetStacksInput struct {
	model.Pagination
	Keyword string `json:"keyword" form:"keyword"`
}
