package project

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type GetListProjectInput struct {
	model.Pagination

	Name   string `form:"name" json:"name"`
	Status string `form:"status" json:"status"`
	Type   string `form:"type" json:"type"`
}

type updateAccountStatusBody struct {
	ProjectStatus model.ProjectStatus `json:"status"`
}

func (i *GetListProjectInput) Validate() error {
	if i.Type != "" && !model.ProjectType(i.Type).IsValid() {
		return ErrInvalidProjectType
	}

	if i.Status != "" && !model.ProjectStatus(i.Status).IsValid() {
		return ErrInvalidProjectStatus
	}
	return nil
}

type CreateProjectInput struct {
	Name              string     `form:"name" json:"name" binding:"required"`
	Status            string     `form:"status" json:"status" binding:"required"`
	Type              string     `form:"type" json:"type" binding:"required"`
	AccountManagerID  model.UUID `form:"accountManagerID" json:"accountManagerID" binding:"required"`
	DeliveryManagerID model.UUID `form:"deliveryManagerID" json:"deliveryManagerID"`
	CountryID         string     `form:"countryID" json:"countryID" binding:"required"`
	StartDate         string     `form:"startDate" json:"startDate"`
}

func (i *CreateProjectInput) Validate() error {
	if !model.ProjectType(i.Type).IsValid() {
		return ErrInvalidProjectType
	}

	if !model.ProjectStatus(i.Status).IsValid() {
		return ErrInvalidProjectStatus
	}

	_, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate != "" && err != nil {
		return ErrInvalidStartDate
	}

	return nil
}

func (i *CreateProjectInput) GetStartDate() *time.Time {
	startDate, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate == "" || err != nil {
		return nil
	}

	return &startDate
}

type GetListStaffInput struct {
	model.Pagination

	Status string `form:"status" json:"status"`
}

func (i *GetListStaffInput) Validate() error {
	if i.Status != "" && !model.ProjectMemberStatus(i.Status).IsValid() {
		return ErrInvalidProjectMemberStatus
	}
	return nil
}
