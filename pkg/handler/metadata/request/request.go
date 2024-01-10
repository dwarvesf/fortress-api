package request

import (
	"github.com/dwarvesf/fortress-api/pkg/handler/metadata/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type UpdateStackBody struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Avatar string `json:"avatar"`
} // @name UpdateStackBody

type UpdateStackInput struct {
	ID   string
	Body UpdateStackBody
} // @name UpdateStackInput

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
} // @name CreateStackInput

type UpdatePositionBody struct {
	Name string `json:"name"`
	Code string `json:"code"`
} // @name UpdatePositionBody

type UpdatePositionInput struct {
	ID   string
	Body UpdatePositionBody
} // @name UpdatePositionInput

func (i UpdatePositionInput) Validate() error {
	if i.ID == "" || !model.IsUUIDFromString(i.ID) {
		return errs.ErrInvalidPositionID
	}

	return nil
}

type CreatePositionInput struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
} // @name CreatePositionInput

type GetStacksInput struct {
	model.Pagination
	Keyword string `json:"keyword" form:"keyword"`
}

type GetBankRequest struct {
	ID        string `json:"id" form:"id" `
	Bin       string `json:"bin" form:"bin"`
	SwiftCode string `json:"swiftCode" form:"swiftCode"`
}
