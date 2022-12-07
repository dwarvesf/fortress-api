package project

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/shopspring/decimal"
)

type GetListProjectInput struct {
	model.Pagination

	Name   string `form:"name" json:"name"`
	Status string `form:"status" json:"status"`
	Type   string `form:"type" json:"type"`
}

type UpdateGeneralInfoInput struct {
	Name      string       `form:"name" json:"name" binding:"required"`
	StartDate string       `form:"startDate" json:"startDate"`
	CountryID model.UUID   `form:"countryID" json:"countryID" binding:"required"`
	Stacks    []model.UUID `form:"stacks" json:"stacks"`
}

func (i UpdateGeneralInfoInput) GetStartDate() *time.Time {
	startDate, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate == "" || err != nil {
		return nil
	}

	return &startDate
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
	Name              string              `form:"name" json:"name" binding:"required"`
	Status            string              `form:"status" json:"status" binding:"required"`
	Type              string              `form:"type" json:"type"`
	AccountManagerID  model.UUID          `form:"accountManagerID" json:"accountManagerID" binding:"required"`
	DeliveryManagerID model.UUID          `form:"deliveryManagerID" json:"deliveryManagerID"`
	CountryID         model.UUID          `form:"countryID" json:"countryID" binding:"required"`
	StartDate         string              `form:"startDate" json:"startDate"`
	Members           []AssignMemberInput `form:"members" json:"members"`
	ClientEmail       string              `form:"clientEmail" json:"clientEmail" binding:"email"`
	ProjectEmail      string              `form:"projectEmail" json:"projectEmail" binding:"email"`
}

func (i *CreateProjectInput) Validate() error {
	if i.Type == "" {
		i.Type = model.ProjectTypeDwarves.String()
	}

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

	for _, member := range i.Members {
		if err := member.Validate(); err != nil {
			return err
		}
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

	Status  string `form:"status" json:"status"`
	Preload bool   `json:"preload" form:"preload,default=true"`
}

func (i *GetListStaffInput) Validate() error {
	if i.Status != "" && !model.ProjectMemberStatus(i.Status).IsValid() {
		return ErrInvalidProjectMemberStatus
	}
	return nil
}

type UpdateMemberInput struct {
	ProjectSlotID  model.UUID      `from:"projectSlotID" json:"projectSlotID" binding:"required"`
	EmployeeID     model.UUID      `form:"employeeID" json:"employeeID"`
	SeniorityID    model.UUID      `form:"seniorityID" json:"seniorityID" binding:"required"`
	Positions      []model.UUID    `form:"positions" json:"positions" binding:"required"`
	DeploymentType string          `form:"deploymentType" json:"deploymentType" binding:"required"`
	Status         string          `form:"status" json:"status" binding:"required"`
	JoinedDate     string          `form:"joinedDate" json:"joinedDate"`
	LeftDate       string          `form:"leftDate" json:"leftDate"`
	Rate           decimal.Decimal `form:"rate" json:"rate" binding:"required"`
	Discount       decimal.Decimal `form:"discount" json:"discount"`
	IsLead         bool            `form:"isLead" json:"isLead"`
}

func (i *UpdateMemberInput) Validate() error {
	if i.DeploymentType != "" && !model.DeploymentType(i.DeploymentType).IsValid() {
		return ErrInvalidDeploymentType
	}

	if i.Status != "" && !model.ProjectMemberStatus(i.Status).IsValid() {
		return ErrInvalidProjectMemberStatus
	}

	_, err := time.Parse("2006-01-02", i.JoinedDate)
	if i.JoinedDate != "" && err != nil {
		return ErrInvalidJoinedDate
	}

	_, err = time.Parse("2006-01-02", i.LeftDate)
	if i.LeftDate != "" && err != nil {
		return ErrInvalidLeftDate
	}

	return nil
}

func (i *UpdateMemberInput) GetJoinedDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.JoinedDate)
	if i.JoinedDate == "" || err != nil {
		return nil
	}

	return &date
}

func (i *UpdateMemberInput) GetLeftDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.LeftDate)
	if i.LeftDate == "" || err != nil {
		return nil
	}

	return &date
}

type AssignMemberInput struct {
	EmployeeID     model.UUID      `form:"employeeID" json:"employeeID"`
	SeniorityID    model.UUID      `form:"seniorityID" json:"seniorityID" binding:"required"`
	Positions      []model.UUID    `form:"positions" json:"positions" binding:"required"`
	DeploymentType string          `form:"deploymentType" json:"deploymentType" binding:"required"`
	Status         string          `form:"status" json:"status" binding:"required"`
	JoinedDate     string          `form:"joinedDate" json:"joinedDate"`
	LeftDate       string          `form:"leftDate" json:"leftDate"`
	Rate           decimal.Decimal `form:"rate" json:"rate" binding:"required"`
	Discount       decimal.Decimal `form:"discount" json:"discount"`
	IsLead         bool            `form:"isLead" json:"isLead"`
}

func (i *AssignMemberInput) Validate() error {
	if i.DeploymentType == "" || !model.DeploymentType(i.DeploymentType).IsValid() {
		return ErrInvalidDeploymentType
	}

	if i.Status == "" ||
		!model.ProjectMemberStatus(i.Status).IsValid() ||
		i.Status == model.ProjectMemberStatusInactive.String() {

		return ErrInvalidProjectMemberStatus
	}

	if len(i.Positions) == 0 {
		return ErrPositionsIsEmpty
	}

	_, err := time.Parse("2006-01-02", i.JoinedDate)
	if i.JoinedDate != "" && err != nil {
		return ErrInvalidJoinedDate
	}

	_, err = time.Parse("2006-01-02", i.LeftDate)
	if i.LeftDate != "" && err != nil {
		return ErrInvalidLeftDate
	}

	if i.Status == model.ProjectMemberStatusPending.String() && !i.EmployeeID.IsZero() {
		i.Status = model.ProjectStatusActive.String()
	}

	return nil
}

func (i *AssignMemberInput) GetJoinedDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.JoinedDate)
	if i.JoinedDate == "" || err != nil {
		return nil
	}

	return &date
}

func (i *AssignMemberInput) GetLeftDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.LeftDate)
	if i.LeftDate == "" || err != nil {
		return nil
	}

	return &date
}

type DeleteMemberInput struct {
	ProjectID string
	MemberID  string
}

func (input DeleteMemberInput) Validate() error {
	if input.ProjectID == "" {
		return ErrInvalidProjectID
	}

	if input.MemberID == "" {
		return ErrInvalidMemberID
	}

	return nil
}

type UpdateContactInfoInput struct {
	ClientEmail       string     `form:"clientEmail" json:"clientEmail" binding:"email"`
	ProjectEmail      string     `form:"projectEmail" json:"projectEmail" binding:"email"`
	AccountManagerID  model.UUID `form:"accountManagerID" json:"accountManagerID" binding:"required"`
	DeliveryManagerID model.UUID `form:"deliveryManagerID" json:"deliveryManagerID"`
}

type UnassignMemberInput struct {
	ProjectID string
	MemberID  string
}

func (input UnassignMemberInput) Validate() error {
	if input.ProjectID == "" || !model.IsUUIDFromString(input.ProjectID) {
		return ErrInvalidProjectID
	}

	if input.MemberID == "" || !model.IsUUIDFromString(input.MemberID) {
		return ErrInvalidMemberID
	}

	return nil
}

type CreateWorkUnitInput struct {
	ProjectID string
	Body      CreateWorkUnitBody
}

type CreateWorkUnitBody struct {
	Name    string       `json:"name" form:"name" binding:"required"`
	Type    string       `json:"type" form:"type" binding:"required"`
	Status  string       `json:"status" form:"status" binding:"required"`
	Members []model.UUID `json:"members" form:"members"`
	Stacks  []model.UUID `json:"stacks" form:"stacks" binding:"required"`
	URL     string       `json:"url" form:"url"`
}

func (i *CreateWorkUnitInput) Validate() error {
	if i.ProjectID == "" || !model.IsUUIDFromString(i.ProjectID) {
		return ErrInvalidProjectID
	}

	return i.Body.Validate()
}

func (i *CreateWorkUnitBody) Validate() error {
	if i.Type == "" || !model.WorkUnitType(i.Type).IsValid() {
		return ErrInvalidWorkUnitType
	}

	if i.Status == "" || !model.WorkUnitStatus(i.Status).IsValid() {
		return ErrInvalidWorkUnitStatus
	}

	if len(i.Stacks) == 0 {
		return ErrInvalidWorkUnitStacks
	}

	return nil
}

type UpdateWorkUnitInput struct {
	ProjectID  string
	WorkUnitID string
	Body       UpdateWorkUnitBody
}

type UpdateWorkUnitBody struct {
	Name    string             `form:"name" json:"name" binding:"required,max=100"`
	Type    model.WorkUnitType `form:"type" json:"type" binding:"required"`
	Members []model.UUID       `form:"members" json:"members"`
	Stacks  []model.UUID       `form:"stacks" json:"stacks" binding:"required"`
	URL     string             `form:"url" json:"url"`
}

func (i *UpdateWorkUnitInput) Validate() error {
	if i.ProjectID == "" || !model.IsUUIDFromString(i.ProjectID) {
		return ErrInvalidProjectID
	}

	if i.WorkUnitID == "" || !model.IsUUIDFromString(i.WorkUnitID) {
		return ErrInvalidWorkUnitID
	}

	return i.Body.Validate()
}

func (i *UpdateWorkUnitBody) Validate() error {
	if !i.Type.IsValid() {
		return ErrInvalidWorkUnitType
	}

	if len(i.Stacks) == 0 {
		return ErrInvalidWorkUnitStacks
	}

	return nil
}

type ArchiveWorkUnitInput struct {
	ProjectID  string
	WorkUnitID string
}

func (i *ArchiveWorkUnitInput) Validate() error {
	if i.ProjectID == "" || !model.IsUUIDFromString(i.ProjectID) {
		return ErrInvalidProjectID
	}

	if i.WorkUnitID == "" || !model.IsUUIDFromString(i.WorkUnitID) {
		return ErrInvalidWorkUnitID
	}

	return nil
}

type GetListWorkUnitInput struct {
	ProjectID string
	Query     GetListWorkUnitQuery
}

type GetListWorkUnitQuery struct {
	Status model.WorkUnitStatus `form:"status" json:"status"`
}

func (i GetListWorkUnitInput) Validate() error {
	if i.ProjectID == "" || !model.IsUUIDFromString(i.ProjectID) {
		return ErrInvalidProjectID
	}

	return i.Query.Validate()
}

func (i GetListWorkUnitQuery) Validate() error {
	if i.Status != "" && !i.Status.IsValid() {
		return ErrInvalidWorkUnitStatus
	}

	return nil
}
