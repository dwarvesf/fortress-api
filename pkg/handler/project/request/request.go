package request

import (
	"regexp"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
	"github.com/dwarvesf/fortress-api/pkg/view"

	"github.com/dwarvesf/fortress-api/pkg/handler/project/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/shopspring/decimal"
)

const emailRegex = ".+@.+\\..+"

type GetListProjectInput struct {
	view.Pagination

	Name   string   `form:"name" json:"name"`
	Status []string `form:"status" json:"status"`
	Type   []string `form:"type" json:"type"`
} // @name GetListProjectInput

type UpdateProjectGeneralInfoRequest struct {
	Name           string      `form:"name" json:"name" binding:"required"`
	StartDate      string      `form:"startDate" json:"startDate"`
	CountryID      view.UUID   `form:"countryID" json:"countryID" binding:"required"`
	Function       string      `form:"function" json:"function" binding:"required"`
	AuditNotionID  view.UUID   `form:"auditNotionID" json:"auditNotionID"`
	Stacks         []view.UUID `form:"stacks" json:"stacks"`
	BankAccountID  view.UUID   `form:"bankAccountID" json:"bankAccountID"`
	ClientID       view.UUID   `form:"clientID" json:"clientID"`
	OrganizationID view.UUID   `form:"organizationID" json:"organizationID"`
	AccountRating  int         `form:"accountRating" json:"accountRating" binding:"required,min=1,max=5"`
	DeliveryRating int         `form:"deliveryRating" json:"deliveryRating" binding:"required,min=1,max=5"`
	LeadRating     int         `form:"leadRating" json:"leadRating" binding:"required,min=1,max=5"`
	ImportantLevel string      `form:"importantLevel" json:"importantLevel" binding:"required"`
} // @name UpdateProjectGeneralInfoRequest

func (i UpdateProjectGeneralInfoRequest) GetStartDate() *time.Time {
	startDate, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate == "" || err != nil {
		return nil
	}

	return &startDate
}

func (i UpdateProjectGeneralInfoRequest) Validate() error {
	if !model.ProjectFunction(i.Function).IsValid() {
		return errs.ErrInvalidProjectFunction
	}

	if !model.ProjectImportantLevel(i.ImportantLevel).IsValid() {
		return errs.ErrInvalidProjectImportantLevel
	}

	return nil
}

type UpdateProjectStatusBody struct {
	ProjectStatus ProjectStatus `json:"status"`
} // @name UpdateProjectStatusBody

type ProjectStatus string // @name ProjectStatus

const (
	ProjectStatusOnBoarding ProjectStatus = "on-boarding"
	ProjectStatusActive     ProjectStatus = "active"
	ProjectStatusPaused     ProjectStatus = "paused"
	ProjectStatusClosed     ProjectStatus = "closed"
)

func (e ProjectStatus) IsValid() bool {
	switch e {
	case
		ProjectStatusOnBoarding,
		ProjectStatusActive,
		ProjectStatusPaused,
		ProjectStatusClosed:
		return true
	}
	return false
}

func (e ProjectStatus) String() string {
	return string(e)
}

func (i *GetListProjectInput) StandardizeInput() {
	statuses := utils.RemoveEmptyString(i.Status)
	pagination := model.Pagination{
		Page: i.Page,
		Size: i.Size,
		Sort: i.Sort,
	}
	pagination.Standardize()
	i.Page = pagination.Page
	i.Size = pagination.Size
	i.Sort = pagination.Sort
	i.Status = statuses
}

func (i *GetListProjectInput) Validate() error {
	if len(i.Type) > 0 {
		for _, projectType := range i.Type {
			if utils.RemoveAllSpace(projectType) != "" && !model.ProjectType(projectType).IsValid() {
				return errs.ErrInvalidProjectType
			}
		}
	}
	if len(i.Status) > 0 {
		for _, status := range i.Status {
			if utils.RemoveAllSpace(status) != "" && !model.ProjectStatus(status).IsValid() {
				return errs.ErrInvalidProjectStatus
			}
		}
	}

	return nil
}

type CreateProjectRequest struct {
	Name             string                `form:"name" json:"name" binding:"required"`
	Status           string                `form:"status" json:"status" binding:"required"`
	Type             string                `form:"type" json:"type"`
	AccountManagers  []ProjectHeadRequest  `form:"accountManagers" json:"accountManagers"`
	DeliveryManagers []ProjectHeadRequest  `form:"deliveryManagers" json:"deliveryManagers"`
	SalePersons      []ProjectHeadRequest  `form:"salePersons" json:"salePersons"`
	CountryID        view.UUID             `form:"countryID" json:"countryID" binding:"required"`
	StartDate        string                `form:"startDate" json:"startDate"`
	Members          []AssignMemberRequest `form:"members" json:"members"`
	ClientEmail      []string              `form:"clientEmail" json:"clientEmail"`
	ProjectEmail     string                `form:"projectEmail" json:"projectEmail"`
	Code             string                `form:"code" json:"code"`
	Function         string                `form:"function" json:"function" binding:"required"`
	AuditNotionID    view.UUID             `form:"auditNotionID" json:"auditNotionID"`
	BankAccountID    view.UUID             `form:"bankAccountID" json:"bankAccountID"`
	ClientID         view.UUID             `form:"clientID" json:"clientID"`
	OrganizationID   view.UUID             `form:"organizationID" json:"organizationID"`
} // @name CreateProjectRequest

func (i *CreateProjectRequest) Validate() error {
	if i.Type == "" {
		i.Type = model.ProjectTypeDwarves.String()
	}

	if !model.ProjectType(i.Type).IsValid() {
		return errs.ErrInvalidProjectType
	}

	if !model.ProjectStatus(i.Status).IsValid() {
		return errs.ErrInvalidProjectStatus
	}

	_, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate != "" && err != nil {
		return errs.ErrInvalidStartDate
	}

	for _, member := range i.Members {
		if err := member.Validate(); err != nil {
			return err
		}
	}

	regex, _ := regexp.Compile(emailRegex)
	for _, v := range i.ClientEmail {
		if !regex.MatchString(v) {
			return errs.ErrInvalidEmailDomainForClient
		}
	}

	if i.ProjectEmail != "" && !regex.MatchString(i.ProjectEmail) {
		return errs.ErrInvalidEmailDomainForProject
	}

	if !model.ProjectFunction(i.Function).IsValid() {
		return errs.ErrInvalidProjectFunction
	}

	if len(i.AccountManagers) == 0 {
		return errs.ErrAccountManagerRequired
	}

	return nil
}

func (i *CreateProjectRequest) GetStartDate() *time.Time {
	startDate, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate == "" || err != nil {
		return nil
	}

	return &startDate
}

type GetListStaffInput struct {
	view.Pagination

	Status   string `form:"status" json:"status"`
	Preload  bool   `json:"preload" form:"preload,default=true"`
	Distinct bool   `json:"distinct" form:"distinct,default=false"`
}

func (i *GetListStaffInput) Validate() error {
	if i.Status != "" && !model.ProjectMemberStatus(i.Status).IsValid() {
		return errs.ErrInvalidProjectMemberStatus
	}
	return nil
}

type UpdateMemberRequest struct {
	ProjectSlotID        view.UUID       `from:"projectSlotID" json:"projectSlotID" binding:"required"`
	ProjectMemberID      view.UUID       `from:"projectMemberID" json:"projectMemberID"`
	EmployeeID           view.UUID       `form:"employeeID" json:"employeeID"`
	SeniorityID          view.UUID       `form:"seniorityID" json:"seniorityID" binding:"required"`
	UpsellPersonID       view.UUID       `form:"upsellPersonID" json:"upsellPersonID"`
	UpsellCommissionRate decimal.Decimal `form:"upsellCommissionRate" json:"upsellCommissionRate"`
	LeadCommissionRate   decimal.Decimal `form:"leadCommissionRate" json:"leadCommissionRate"`
	Positions            []view.UUID     `form:"positions" json:"positions" binding:"required"`
	DeploymentType       string          `form:"deploymentType" json:"deploymentType" binding:"required"`
	Status               string          `form:"status" json:"status" binding:"required"`
	StartDate            string          `form:"startDate" json:"startDate"`
	EndDate              string          `form:"endDate" json:"endDate"`
	Rate                 decimal.Decimal `form:"rate" json:"rate" binding:"required"`
	Discount             decimal.Decimal `form:"discount" json:"discount"`
	IsLead               bool            `form:"isLead" json:"isLead"`
	Note                 string          `form:"note" json:"note"`
} // @name UpdateMemberRequest

func (i *UpdateMemberRequest) Validate() error {
	if i.DeploymentType != "" && !model.DeploymentType(i.DeploymentType).IsValid() {
		return errs.ErrInvalidDeploymentType
	}

	if i.Status != "" && !model.ProjectMemberStatus(i.Status).IsValid() {
		return errs.ErrInvalidProjectMemberStatus
	}

	_, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate != "" && err != nil {
		return errs.ErrInvalidStartDate
	}

	_, err = time.Parse("2006-01-02", i.EndDate)
	if i.EndDate != "" && err != nil {
		return errs.ErrInvalidEndDate
	}

	if i.GetStartDate() != nil &&
		i.GetEndDate() != nil &&
		!i.GetStartDate().Before(*i.GetEndDate()) {
		return errs.ErrInvalidEndDate
	}

	if i.GetEndDate() != nil && i.GetEndDate().Before(time.Now()) {
		i.Status = model.ProjectMemberStatusInactive.String()
	}

	return nil
}

func (i *UpdateMemberRequest) GetStartDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate == "" || err != nil {
		return nil
	}

	return &date
}

func (i *UpdateMemberRequest) GetEndDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.EndDate)
	if i.EndDate == "" || err != nil {
		return nil
	}

	return &date
}

type AssignMemberRequest struct {
	EmployeeID           view.UUID       `form:"employeeID" json:"employeeID"`
	SeniorityID          view.UUID       `form:"seniorityID" json:"seniorityID" binding:"required"`
	Positions            []view.UUID     `form:"positions" json:"positions" binding:"required"`
	DeploymentType       string          `form:"deploymentType" json:"deploymentType" binding:"required"`
	Status               string          `form:"status" json:"status" binding:"required"`
	StartDate            string          `form:"startDate" json:"startDate"`
	EndDate              string          `form:"endDate" json:"endDate"`
	Rate                 decimal.Decimal `form:"rate" json:"rate" binding:"required"`
	Discount             decimal.Decimal `form:"discount" json:"discount"`
	LeadCommissionRate   decimal.Decimal `form:"leadCommissionRate" json:"leadCommissionRate"`
	IsLead               bool            `form:"isLead" json:"isLead"`
	UpsellPersonID       view.UUID       `form:"upsellPersonID" json:"upsellPersonID"`
	UpsellCommissionRate decimal.Decimal `form:"upsellCommissionRate" json:"upsellCommissionRate"`
	Note                 string          `form:"note" json:"note"`
} // @name AssignMemberRequest

func (i *AssignMemberRequest) Validate() error {
	if i.DeploymentType == "" || !model.DeploymentType(i.DeploymentType).IsValid() {
		return errs.ErrInvalidDeploymentType
	}

	if i.Status == "" ||
		!model.ProjectMemberStatus(i.Status).IsValid() ||
		i.Status == model.ProjectMemberStatusInactive.String() {
		return errs.ErrInvalidProjectMemberStatus
	}

	if len(i.Positions) == 0 {
		return errs.ErrPositionsIsEmpty
	}

	_, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate != "" && err != nil {
		return errs.ErrInvalidStartDate
	}

	_, err = time.Parse("2006-01-02", i.EndDate)
	if i.EndDate != "" && err != nil {
		return errs.ErrInvalidEndDate
	}

	if i.Status == model.ProjectMemberStatusPending.String() && !i.EmployeeID.IsZero() {
		i.Status = model.ProjectStatusActive.String()
	}

	return nil
}

func (i *AssignMemberRequest) GetStartDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate == "" || err != nil {
		return nil
	}

	return &date
}

func (i *AssignMemberRequest) GetEndDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.EndDate)
	if i.EndDate == "" || err != nil {
		return nil
	}

	return &date
}

func (i *AssignMemberRequest) GetStatus() model.ProjectMemberStatus {
	if i.EmployeeID.IsZero() {
		return model.ProjectMemberStatusPending
	}

	if !i.EmployeeID.IsZero() && i.Status == model.ProjectMemberStatusPending.String() {
		return model.ProjectMemberStatusActive
	}

	return model.ProjectMemberStatus(i.Status)
}

func (i *AssignMemberRequest) RestrictPermission(userInfo *model.CurrentLoggedUserInfo) {
	if !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateEdit) {
		i.LeadCommissionRate = decimal.Zero
	}

	if !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectMembersRateEdit) {
		i.Rate = decimal.Zero
		i.Discount = decimal.Zero
	}
}

type DeleteMemberInput struct {
	ProjectID string
	MemberID  string
}

func (input DeleteMemberInput) Validate() error {
	if input.ProjectID == "" {
		return errs.ErrInvalidProjectID
	}

	if input.MemberID == "" {
		return errs.ErrInvalidMemberID
	}

	return nil
}

type DeleteSlotInput struct {
	ProjectID string
	SlotID    string
}

func (input DeleteSlotInput) Validate() error {
	if input.ProjectID == "" {
		return errs.ErrInvalidProjectID
	}

	if input.SlotID == "" {
		return errs.ErrInvalidSlotID
	}

	return nil
}

type ProjectHeadRequest struct {
	EmployeeID     view.UUID       `json:"employeeID" form:"employeeID"`
	CommissionRate decimal.Decimal `json:"commissionRate" form:"commissionRate"`
} // @name ProjectHeadRequest

type UpdateContactInfoRequest struct {
	ClientEmail      []string             `form:"clientEmail" json:"clientEmail"`
	ProjectEmail     string               `form:"projectEmail" json:"projectEmail"`
	AccountManagers  []ProjectHeadRequest `form:"accountManagers" json:"accountManagers"`
	DeliveryManagers []ProjectHeadRequest `form:"deliveryManagers" json:"deliveryManagers"`
	SalePersons      []ProjectHeadRequest `form:"salePersons" json:"salePersons"`
} // @name UpdateContactInfoRequest

func (i UpdateContactInfoRequest) Validate() error {
	regex, _ := regexp.Compile(emailRegex)
	for _, v := range i.ClientEmail {
		if !regex.MatchString(v) {
			return errs.ErrInvalidEmailDomainForClient
		}
	}

	if i.ProjectEmail != "" && !regex.MatchString(i.ProjectEmail) {
		return errs.ErrInvalidEmailDomainForProject
	}

	if len(i.AccountManagers) == 0 {
		return errs.ErrAccountManagerRequired
	}

	return nil
}

type UnassignMemberInput struct {
	ProjectID string
	MemberID  string
}

func (input UnassignMemberInput) Validate() error {
	if input.ProjectID == "" || !model.IsUUIDFromString(input.ProjectID) {
		return errs.ErrInvalidProjectID
	}

	if input.MemberID == "" || !model.IsUUIDFromString(input.MemberID) {
		return errs.ErrInvalidMemberID
	}

	return nil
}

type CreateWorkUnitInput struct {
	ProjectID string
	Body      CreateWorkUnitRequest
}

type CreateWorkUnitRequest struct {
	Name    string      `json:"name" form:"name" binding:"required"`
	Type    string      `json:"type" form:"type" binding:"required"`
	Status  string      `json:"status" form:"status" binding:"required"`
	Members []view.UUID `json:"members" form:"members"`
	Stacks  []view.UUID `json:"stacks" form:"stacks" binding:"required"`
	URL     string      `json:"url" form:"url"`
} // @name CreateWorkUnitRequest

func (i *CreateWorkUnitInput) Validate() error {
	if i.ProjectID == "" || !model.IsUUIDFromString(i.ProjectID) {
		return errs.ErrInvalidProjectID
	}

	return i.Body.Validate()
}

func (i *CreateWorkUnitRequest) Validate() error {
	if i.Type == "" || !model.WorkUnitType(i.Type).IsValid() {
		return errs.ErrInvalidWorkUnitType
	}

	if i.Status == "" || !model.WorkUnitStatus(i.Status).IsValid() {
		return errs.ErrInvalidWorkUnitStatus
	}

	if len(i.Stacks) == 0 {
		return errs.ErrInvalidWorkUnitStacks
	}

	return nil
}

type UpdateWorkUnitInput struct {
	ProjectID  string
	WorkUnitID string
	Body       UpdateWorkUnitRequest
}

type UpdateWorkUnitRequest struct {
	Name    string            `form:"name" json:"name" binding:"required,max=100"`
	Type    view.WorkUnitType `form:"type" json:"type" binding:"required"`
	Members []view.UUID       `form:"members" json:"members"`
	Stacks  []view.UUID       `form:"stacks" json:"stacks" binding:"required"`
	URL     string            `form:"url" json:"url"`
} // @name UpdateWorkUnitRequest

func (i *UpdateWorkUnitInput) Validate() error {
	if i.ProjectID == "" || !model.IsUUIDFromString(i.ProjectID) {
		return errs.ErrInvalidProjectID
	}

	if i.WorkUnitID == "" || !model.IsUUIDFromString(i.WorkUnitID) {
		return errs.ErrInvalidWorkUnitID
	}

	return i.Body.Validate()
}

func (i *UpdateWorkUnitRequest) Validate() error {
	if !i.Type.IsValid() {
		return errs.ErrInvalidWorkUnitType
	}

	if len(i.Stacks) == 0 {
		return errs.ErrInvalidWorkUnitStacks
	}

	return nil
}

type ArchiveWorkUnitInput struct {
	ProjectID  string
	WorkUnitID string
}

func (i *ArchiveWorkUnitInput) Validate() error {
	if i.ProjectID == "" || !model.IsUUIDFromString(i.ProjectID) {
		return errs.ErrInvalidProjectID
	}

	if i.WorkUnitID == "" || !model.IsUUIDFromString(i.WorkUnitID) {
		return errs.ErrInvalidWorkUnitID
	}

	return nil
}

type GetListWorkUnitInput struct {
	ProjectID string
	Query     GetListWorkUnitQuery
} // @name GetListWorkUnitInput

type GetListWorkUnitQuery struct {
	Status WorkUnitStatus `form:"status" json:"status"`
} // @name GetListWorkUnitQuery

type WorkUnitStatus string // @name WorkUnitStatus

const (
	WorkUnitStatusActive   WorkUnitStatus = "active"
	WorkUnitStatusArchived WorkUnitStatus = "archived"
)

func (e WorkUnitStatus) IsValid() bool {
	switch e {
	case
		WorkUnitStatusActive,
		WorkUnitStatusArchived:
		return true
	}
	return false
}

func (e WorkUnitStatus) String() string {
	return string(e)
}

func (i GetListWorkUnitInput) Validate() error {
	if i.ProjectID == "" || !model.IsUUIDFromString(i.ProjectID) {
		return errs.ErrInvalidProjectID
	}

	return i.Query.Validate()
}

func (i GetListWorkUnitQuery) Validate() error {
	if i.Status != "" && !i.Status.IsValid() {
		return errs.ErrInvalidWorkUnitStatus
	}

	return nil
}

type UpdateSendingSurveyInput struct {
	AllowsSendingSurvey bool `form:"allowsSendingSurvey" json:"allowsSendingSurvey"`
}
