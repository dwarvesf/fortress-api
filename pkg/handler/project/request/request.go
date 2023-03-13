package request

import (
	"regexp"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/utils"

	"github.com/dwarvesf/fortress-api/pkg/handler/project/errs"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/shopspring/decimal"
)

const emailRegex = ".+@.+\\..+"

type GetListProjectInput struct {
	model.Pagination

	Name   string   `form:"name" json:"name"`
	Status []string `form:"status" json:"status"`
	Type   string   `form:"type" json:"type"`
}

type UpdateProjectGeneralInfoInput struct {
	Name           string       `form:"name" json:"name" binding:"required"`
	StartDate      string       `form:"startDate" json:"startDate"`
	CountryID      model.UUID   `form:"countryID" json:"countryID" binding:"required"`
	Function       string       `form:"function" json:"function" binding:"required"`
	AuditNotionID  model.UUID   `form:"auditNotionID" json:"auditNotionID"`
	Stacks         []model.UUID `form:"stacks" json:"stacks"`
	BankAccountID  model.UUID   `form:"bankAccountID" json:"bankAccountID"`
	ClientID       model.UUID   `form:"clientID" json:"clientID"`
	OrganizationID model.UUID   `form:"organizationID" json:"organizationID"`
}

func (i UpdateProjectGeneralInfoInput) GetStartDate() *time.Time {
	startDate, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate == "" || err != nil {
		return nil
	}

	return &startDate
}

func (i UpdateProjectGeneralInfoInput) Validate() error {
	if !model.ProjectFunction(i.Function).IsValid() {
		return errs.ErrInvalidProjectFunction
	}

	return nil
}

type UpdateAccountStatusBody struct {
	ProjectStatus model.ProjectStatus `json:"status"`
}

func (i *GetListProjectInput) StandardizeInput() {
	statuses := utils.RemoveEmptyString(i.Status)
	i.Pagination.Standardize()
	i.Status = statuses
}

func (i *GetListProjectInput) Validate() error {
	if i.Type != "" && !model.ProjectType(i.Type).IsValid() {
		return errs.ErrInvalidProjectType
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

type CreateProjectInput struct {
	Name             string              `form:"name" json:"name" binding:"required"`
	Status           string              `form:"status" json:"status" binding:"required"`
	Type             string              `form:"type" json:"type"`
	AccountManagers  []ProjectHeadInput  `form:"accountManagers" json:"accountManagers"`
	DeliveryManagers []ProjectHeadInput  `form:"deliveryManagers" json:"deliveryManagers"`
	SalePersons      []ProjectHeadInput  `form:"salePersons" json:"salePersons"`
	CountryID        model.UUID          `form:"countryID" json:"countryID" binding:"required"`
	StartDate        string              `form:"startDate" json:"startDate"`
	Members          []AssignMemberInput `form:"members" json:"members"`
	ClientEmail      []string            `form:"clientEmail" json:"clientEmail"`
	ProjectEmail     string              `form:"projectEmail" json:"projectEmail"`
	Code             string              `form:"code" json:"code"`
	Function         string              `form:"function" json:"function" binding:"required"`
	AuditNotionID    model.UUID          `form:"auditNotionID" json:"auditNotionID"`
	BankAccountID    model.UUID          `form:"bankAccountID" json:"bankAccountID"`
	ClientID         model.UUID          `form:"clientID" json:"clientID"`
	OrganizationID   model.UUID          `form:"organizationID" json:"organizationID"`
}

func (i *CreateProjectInput) Validate() error {
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

func (i *CreateProjectInput) GetStartDate() *time.Time {
	startDate, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate == "" || err != nil {
		return nil
	}

	return &startDate
}

type GetListStaffInput struct {
	model.Pagination

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

type UpdateMemberInput struct {
	ProjectSlotID        model.UUID      `from:"projectSlotID" json:"projectSlotID" binding:"required"`
	ProjectMemberID      model.UUID      `from:"projectMemberID" json:"projectMemberID"`
	EmployeeID           model.UUID      `form:"employeeID" json:"employeeID"`
	SeniorityID          model.UUID      `form:"seniorityID" json:"seniorityID" binding:"required"`
	UpsellPersonID       model.UUID      `form:"upsellPersonID" json:"upsellPersonID"`
	UpsellCommissionRate decimal.Decimal `form:"upsellCommissionRate" json:"upsellCommissionRate"`
	LeadCommissionRate   decimal.Decimal `form:"leadCommissionRate" json:"leadCommissionRate"`
	Positions            []model.UUID    `form:"positions" json:"positions" binding:"required"`
	DeploymentType       string          `form:"deploymentType" json:"deploymentType" binding:"required"`
	Status               string          `form:"status" json:"status" binding:"required"`
	StartDate            string          `form:"startDate" json:"startDate"`
	EndDate              string          `form:"endDate" json:"endDate"`
	Rate                 decimal.Decimal `form:"rate" json:"rate" binding:"required"`
	Discount             decimal.Decimal `form:"discount" json:"discount"`
	IsLead               bool            `form:"isLead" json:"isLead"`
	Note                 string          `form:"note" json:"note"`
}

func (i *UpdateMemberInput) Validate() error {
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

func (i *UpdateMemberInput) GetStartDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate == "" || err != nil {
		return nil
	}

	return &date
}

func (i *UpdateMemberInput) GetEndDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.EndDate)
	if i.EndDate == "" || err != nil {
		return nil
	}

	return &date
}

type AssignMemberInput struct {
	EmployeeID         model.UUID      `form:"employeeID" json:"employeeID"`
	SeniorityID        model.UUID      `form:"seniorityID" json:"seniorityID" binding:"required"`
	Positions          []model.UUID    `form:"positions" json:"positions" binding:"required"`
	DeploymentType     string          `form:"deploymentType" json:"deploymentType" binding:"required"`
	Status             string          `form:"status" json:"status" binding:"required"`
	StartDate          string          `form:"startDate" json:"startDate"`
	EndDate            string          `form:"endDate" json:"endDate"`
	Rate               decimal.Decimal `form:"rate" json:"rate" binding:"required"`
	Discount           decimal.Decimal `form:"discount" json:"discount"`
	LeadCommissionRate decimal.Decimal `form:"leadCommissionRate" json:"leadCommissionRate"`
	IsLead             bool            `form:"isLead" json:"isLead"`
	UpsellPersonID     model.UUID      `form:"upsellPersonID" json:"upsellPersonID"`
	Note               string          `form:"note" json:"note"`
}

func (i *AssignMemberInput) Validate() error {
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

func (i *AssignMemberInput) GetStartDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.StartDate)
	if i.StartDate == "" || err != nil {
		return nil
	}

	return &date
}

func (i *AssignMemberInput) GetEndDate() *time.Time {
	date, err := time.Parse("2006-01-02", i.EndDate)
	if i.EndDate == "" || err != nil {
		return nil
	}

	return &date
}

func (i *AssignMemberInput) GetStatus() model.ProjectMemberStatus {
	if i.EmployeeID.IsZero() {
		return model.ProjectMemberStatusPending
	}

	if !i.EmployeeID.IsZero() && i.Status == model.ProjectMemberStatusPending.String() {
		return model.ProjectMemberStatusActive
	}

	return model.ProjectMemberStatus(i.Status)
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

type ProjectHeadInput struct {
	EmployeeID     model.UUID `json:"employeeID" form:"employeeID"`
	CommissionRate int        `json:"commissionRate" form:"commissionRate"`
}

type UpdateContactInfoInput struct {
	ClientEmail      []string           `form:"clientEmail" json:"clientEmail"`
	ProjectEmail     string             `form:"projectEmail" json:"projectEmail"`
	AccountManagers  []ProjectHeadInput `form:"accountManagers" json:"accountManagers"`
	DeliveryManagers []ProjectHeadInput `form:"deliveryManagers" json:"deliveryManagers"`
	SalePersons      []ProjectHeadInput `form:"salePersons" json:"salePersons"`
}

func (i UpdateContactInfoInput) Validate() error {
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

	if !isValidCommissionRate(i.AccountManagers) ||
		!isValidCommissionRate(i.DeliveryManagers) ||
		!isValidCommissionRate(i.SalePersons) {
		return errs.ErrTotalCommissionRateMustBe100
	}

	return nil
}

func isValidCommissionRate(heads []ProjectHeadInput) bool {
	if len(heads) == 0 {
		return true
	}

	sum := 0
	for _, head := range heads {
		sum += head.CommissionRate
	}
	return sum == 100
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
		return errs.ErrInvalidProjectID
	}

	return i.Body.Validate()
}

func (i *CreateWorkUnitBody) Validate() error {
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
		return errs.ErrInvalidProjectID
	}

	if i.WorkUnitID == "" || !model.IsUUIDFromString(i.WorkUnitID) {
		return errs.ErrInvalidWorkUnitID
	}

	return i.Body.Validate()
}

func (i *UpdateWorkUnitBody) Validate() error {
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
}

type GetListWorkUnitQuery struct {
	Status model.WorkUnitStatus `form:"status" json:"status"`
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
