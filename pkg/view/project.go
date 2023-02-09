package view

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

type ProjectData struct {
	model.BaseModel

	Name                string                `json:"name"`
	Avatar              string                `json:"avatar"`
	Type                string                `json:"type"`
	Status              string                `json:"status"`
	ProjectEmail        string                `json:"projectEmail"`
	ClientEmail         []string              `json:"clientEmail"`
	Industry            string                `json:"industry"`
	AllowsSendingSurvey bool                  `json:"allowsSendingSurvey"`
	Country             *BasicCountryInfo     `json:"country"`
	StartDate           *time.Time            `json:"startDate"`
	EndDate             *time.Time            `json:"endDate"`
	Members             []ProjectMember       `json:"members"`
	TechnicalLead       []ProjectHead         `json:"technicalLeads"`
	AccountManager      *ProjectHead          `json:"accountManager"`
	SalePerson          *ProjectHead          `json:"salePerson"`
	DeliveryManager     *ProjectHead          `json:"deliveryManager"`
	Stacks              []Stack               `json:"stacks"`
	Code                string                `json:"code"`
	Function            string                `json:"function"`
	AuditNotionID       string                `json:"auditNotionID"`
	BankAccount         *BasicBankAccountInfo `json:"bankAccount"`
	Client              *Client               `json:"client"`
}

type UpdatedProject struct {
	model.BaseModel

	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

type ProjectMember struct {
	ProjectMemberID string          `json:"projectMemberID"`
	ProjectSlotID   string          `json:"projectSlotID"`
	EmployeeID      string          `json:"employeeID"`
	FullName        string          `json:"fullName"`
	DisplayName     string          `json:"displayName"`
	Avatar          string          `json:"avatar"`
	Username        string          `json:"username"`
	Status          string          `json:"status"`
	IsLead          bool            `json:"isLead"`
	DeploymentType  string          `json:"deploymentType"`
	StartDate       *time.Time      `json:"startDate"`
	EndDate         *time.Time      `json:"endDate"`
	Rate            decimal.Decimal `json:"rate"`
	Discount        decimal.Decimal `json:"discount"`

	Seniority *model.Seniority `json:"seniority"`
	Positions []Position       `json:"positions"`
}

type ProjectHead struct {
	EmployeeID  string `json:"employeeID"`
	FullName    string `json:"fullName"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Username    string `json:"username"`
}

func ToProjectHead(head *model.ProjectHead) *ProjectHead {
	return &ProjectHead{
		EmployeeID:  head.EmployeeID.String(),
		FullName:    head.Employee.FullName,
		DisplayName: head.Employee.DisplayName,
		Avatar:      head.Employee.Avatar,
		Username:    head.Employee.Username,
	}
}

type UpdateProjectStatusResponse struct {
	Data UpdatedProject `json:"data"`
}

func ToUpdateProjectStatusResponse(p *model.Project) UpdatedProject {
	return UpdatedProject{
		BaseModel: p.BaseModel,
		Name:      p.Name,
		Type:      p.Type.String(),
		Status:    p.Status.String(),
		StartDate: p.StartDate,
		EndDate:   p.EndDate,
	}
}

func ToProjectData(c *gin.Context, project *model.Project, userInfo *model.CurrentLoggedUserInfo) ProjectData {
	leadMap := map[string]bool{}
	var technicalLeads = make([]ProjectHead, 0, len(project.Heads))
	var accountManager, salePerson, deliveryManager *ProjectHead
	for _, h := range project.Heads {
		head := ToProjectHead(h)

		if h.IsLead() {
			leadMap[h.EmployeeID.String()] = true
			technicalLeads = append(technicalLeads, *head)
			continue
		}

		if h.IsAccountManager() {
			accountManager = head
			continue
		}

		if h.IsSalePerson() {
			salePerson = head
			continue
		}

		if h.IsDeliveryManager() {
			deliveryManager = head
		}
	}

	var members = make([]ProjectMember, 0, len(project.ProjectMembers))
	for _, m := range project.ProjectMembers {
		member := ProjectMember{
			Status:      m.Status.String(),
			EmployeeID:  m.EmployeeID.String(),
			FullName:    m.Employee.FullName,
			DisplayName: m.Employee.DisplayName,
			Avatar:      m.Employee.Avatar,
			Username:    m.Employee.Username,
			Seniority:   m.Seniority,
			IsLead:      leadMap[m.EmployeeID.String()],
			Positions:   ToProjectMemberPositions(m.ProjectMemberPositions),
		}

		if utils.HasPermission(c, userInfo.Permissions, model.PermissionProjectsReadFullAccess) {
			member.DeploymentType = m.DeploymentType.String()
		}

		members = append(members, member)
	}

	d := ProjectData{
		BaseModel:       project.BaseModel,
		Avatar:          project.Avatar,
		Name:            project.Name,
		Type:            project.Type.String(),
		Status:          project.Status.String(),
		Stacks:          ToProjectStacks(project.ProjectStacks),
		StartDate:       project.StartDate,
		EndDate:         project.EndDate,
		Members:         members,
		TechnicalLead:   technicalLeads,
		DeliveryManager: deliveryManager,
		SalePerson:      salePerson,
		AccountManager:  accountManager,
		ProjectEmail:    project.ProjectEmail,

		AllowsSendingSurvey: project.AllowsSendingSurvey,
		Code:                project.Code,
		Function:            project.Function.String(),
	}

	var clientEmail []string
	if project.ClientEmail != "" {
		clientEmail = strings.Split(project.ClientEmail, ",")
	}

	if utils.HasPermission(c, userInfo.Permissions, model.PermissionProjectsReadFullAccess) {
		if project.ProjectNotion != nil && !project.ProjectNotion.AuditNotionID.IsZero() {
			d.AuditNotionID = project.ProjectNotion.AuditNotionID.String()
		}

		d.ClientEmail = clientEmail

		if project.BankAccount != nil {
			d.BankAccount = &BasicBankAccountInfo{
				ID:            project.BankAccount.ID.String(),
				AccountNumber: project.BankAccount.AccountNumber,
				BankName:      project.BankAccount.BankName,
				OwnerName:     project.BankAccount.OwnerName,
			}
		}

		if project.Client != nil {
			d.Client = toClient(project.Client)
		}
	}

	if project.Country != nil {
		d.Country = &BasicCountryInfo{
			ID:   project.Country.ID,
			Name: project.Country.Name,
			Code: project.Country.Code,
		}
	}

	return d
}

func ToProjectsData(c *gin.Context, projects []*model.Project, userInfo *model.CurrentLoggedUserInfo) []ProjectData {
	var results = make([]ProjectData, 0, len(projects))

	for _, p := range projects {
		// If the project belongs user, append it in the list
		_, ok := userInfo.Projects[p.ID]
		if ok && p.Status == model.ProjectStatusActive && model.IsUserActiveInProject(userInfo.UserID, p.ProjectMembers) {
			results = append(results, ToProjectData(c, p, userInfo))
			continue
		}

		// If the project is not belong user, check if the user has permission to view the project
		if utils.HasPermission(c, userInfo.Permissions, model.PermissionProjectsReadFullAccess) ||
			utils.HasPermission(c, userInfo.Permissions, model.PermissionEmployeesReadProjectsReadActive) {

			if p.Status == model.ProjectStatusActive {
				results = append(results, ToProjectData(c, p, userInfo))
			} else {
				if utils.HasPermission(c, userInfo.Permissions, model.PermissionProjectsReadFullAccess) {
					results = append(results, ToProjectData(c, p, userInfo))
				}
			}
		}
	}

	return results
}

type ProjectListDataResponse struct {
	Data []ProjectData `json:"data"`
}

type ProjectDataResponse struct {
	Data ProjectData `json:"data"`
}

type CreateMemberData struct {
	ProjectSlotID   string          `json:"projectSlotID"`
	ProjectMemberID string          `json:"projectMemberID"`
	EmployeeID      string          `json:"employeeID"`
	FullName        string          `json:"fullName"`
	DisplayName     string          `json:"displayName"`
	Avatar          string          `json:"avatar"`
	Positions       []Position      `json:"positions"`
	DeploymentType  string          `json:"deploymentType"`
	Status          string          `json:"status"`
	IsLead          bool            `json:"isLead"`
	Seniority       model.Seniority `json:"seniority"`
	Username        string          `json:"username"`
}

type CreateMemberDataResponse struct {
	Data CreateMemberData `json:"data"`
}

func ToCreateMemberData(slot *model.ProjectSlot) CreateMemberData {
	rs := CreateMemberData{
		ProjectSlotID:  slot.ID.String(),
		FullName:       slot.ProjectMember.Employee.FullName,
		DisplayName:    slot.ProjectMember.Employee.DisplayName,
		Avatar:         slot.ProjectMember.Employee.Avatar,
		Username:       slot.ProjectMember.Employee.Username,
		DeploymentType: slot.DeploymentType.String(),
		Status:         slot.Status.String(),
		Positions:      ToProjectSlotPositions(slot.ProjectSlotPositions),
		IsLead:         slot.ProjectMember.IsLead,
		Seniority:      slot.Seniority,
	}

	if !slot.ProjectMember.ID.IsZero() {
		rs.ProjectMemberID = slot.ProjectMember.ID.String()
		rs.EmployeeID = slot.ProjectMember.EmployeeID.String()
	}

	return rs
}

type CreateProjectData struct {
	model.BaseModel

	Name            string                `json:"name"`
	Type            string                `json:"type"`
	Status          string                `json:"status"`
	StartDate       string                `json:"startDate"`
	AccountManager  *ProjectHead          `json:"accountManager"`
	DeliveryManager *ProjectHead          `json:"deliveryManager"`
	Members         []CreateMemberData    `json:"members"`
	ClientEmail     []string              `json:"clientEmail"`
	ProjectEmail    string                `json:"projectEmail"`
	Country         *BasicCountryInfo     `json:"country"`
	Code            string                `json:"code"`
	Function        string                `json:"function"`
	BankAccount     *BasicBankAccountInfo `json:"bankAccount"`
	Client          *Client               `json:"client"`
}

type BasicBankAccountInfo struct {
	ID            string `json:"id"`
	AccountNumber string `json:"accountNumber"`
	BankName      string `json:"bankName"`
	OwnerName     string `json:"ownerName"`
}

func ToCreateProjectDataResponse(project *model.Project) CreateProjectData {
	var clientEmail []string
	if project.ClientEmail != "" {
		clientEmail = strings.Split(project.ClientEmail, ",")
	}

	result := CreateProjectData{
		BaseModel:    project.BaseModel,
		Name:         project.Name,
		Type:         project.Type.String(),
		Status:       project.Status.String(),
		ClientEmail:  clientEmail,
		ProjectEmail: project.ProjectEmail,
		Code:         project.Code,
		Function:     project.Function.String(),
	}

	if project.Country != nil {
		result.Country = &BasicCountryInfo{
			ID:   project.Country.ID,
			Name: project.Country.Name,
			Code: project.Country.Code,
		}
	}

	if project.BankAccount != nil {
		result.BankAccount = &BasicBankAccountInfo{
			ID:            project.BankAccount.ID.String(),
			AccountNumber: project.BankAccount.AccountNumber,
			BankName:      project.BankAccount.BankName,
			OwnerName:     project.BankAccount.OwnerName,
		}
	}

	if project.Client != nil {
		result.Client = toClient(project.Client)
	}

	if project.StartDate != nil {
		result.StartDate = project.StartDate.Format("2006-01-02")
	}

	for _, head := range project.Heads {
		switch head.Position {
		case model.HeadPositionAccountManager:
			result.AccountManager = ToProjectHead(head)
		case model.HeadPositionDeliveryManager:
			result.DeliveryManager = ToProjectHead(head)
		}
	}

	result.Members = make([]CreateMemberData, 0, len(project.Slots))
	for _, slot := range project.Slots {
		result.Members = append(result.Members, ToCreateMemberData(&slot))
	}

	return result
}

func ToProjectMemberListData(members []*model.ProjectMember, projectHeads []*model.ProjectHead) []ProjectMember {
	var results = make([]ProjectMember, 0, len(members))

	leadMap := map[string]bool{}
	for _, v := range projectHeads {
		if v.IsLead() {
			leadMap[v.EmployeeID.String()] = true
		}
	}

	for _, m := range members {
		var member ProjectMember

		if m.ID.IsZero() {
			member = ProjectMember{
				ProjectSlotID:  m.ProjectSlotID.String(),
				Status:         m.Status.String(),
				DeploymentType: m.DeploymentType.String(),
				Rate:           m.Rate,
				Discount:       m.Discount,
				Seniority:      m.Seniority,
				Positions:      ToPositions(m.Positions),
			}
		} else {
			member = ProjectMember{
				ProjectSlotID:   m.ProjectSlotID.String(),
				ProjectMemberID: m.ID.String(),
				EmployeeID:      m.EmployeeID.String(),
				FullName:        m.Employee.FullName,
				DisplayName:     m.Employee.DisplayName,
				Avatar:          m.Employee.Avatar,
				Username:        m.Employee.Username,
				StartDate:       m.StartDate,
				EndDate:         m.EndDate,
				IsLead:          leadMap[m.EmployeeID.String()],
				Status:          m.Status.String(),
				DeploymentType:  m.DeploymentType.String(),
				Rate:            m.Rate,
				Discount:        m.Discount,
				Seniority:       m.Seniority,
				Positions:       ToProjectMemberPositions(m.ProjectMemberPositions),
			}
		}

		results = append(results, member)
	}

	return results
}

type ProjectMemberListResponse struct {
	Data []ProjectMember `json:"data"`
}

type BasicCountryInfo struct {
	ID   model.UUID `json:"id"`
	Name string     `json:"name"`
	Code string     `json:"code"`
}

type UpdateProjectGeneralInfo struct {
	Name          string                `json:"name"`
	StartDate     *time.Time            `json:"startDate"`
	Country       *BasicCountryInfo     `json:"country"`
	Stacks        []model.Stack         `json:"stacks"`
	Function      model.ProjectFunction `json:"function"`
	AuditNotionID string                `json:"auditNotionID"`
	BankAccount   *BasicBankAccountInfo `json:"bankAccount"`
	Client        *Client               `json:"client"`
}

type UpdateProjectGeneralInfoResponse struct {
	Data UpdateProjectGeneralInfo `json:"data"`
}

func ToUpdateProjectGeneralInfo(project *model.Project) UpdateProjectGeneralInfo {
	stacks := make([]model.Stack, 0, len(project.ProjectStacks))
	for _, v := range project.ProjectStacks {
		stacks = append(stacks, v.Stack)
	}

	rs := UpdateProjectGeneralInfo{
		Name:      project.Name,
		StartDate: project.StartDate,
		Stacks:    stacks,
		Function:  project.Function,
	}

	if project.ProjectNotion != nil && !project.ProjectNotion.AuditNotionID.IsZero() {
		rs.AuditNotionID = project.ProjectNotion.AuditNotionID.String()
	}

	if project.Country != nil {
		rs.Country = &BasicCountryInfo{
			ID:   project.Country.ID,
			Name: project.Country.Name,
			Code: project.Country.Code,
		}
	}

	if project.BankAccount != nil {
		rs.BankAccount = &BasicBankAccountInfo{
			ID:            project.BankAccount.ID.String(),
			AccountNumber: project.BankAccount.AccountNumber,
			BankName:      project.BankAccount.BankName,
			OwnerName:     project.BankAccount.OwnerName,
		}
	}

	if project.Client != nil {
		rs.Client = toClient(project.Client)
	}

	return rs
}

type BasicProjectHeadInfo struct {
	EmployeeID  string             `json:"employeeID"`
	FullName    string             `json:"fullName"`
	DisplayName string             `json:"displayName"`
	Avatar      string             `json:"avatar"`
	Position    model.HeadPosition `json:"position"`
	Username    string             `json:"username"`
}

type UpdateProjectContactInfo struct {
	ClientEmail  []string               `json:"clientEmail"`
	ProjectEmail string                 `json:"projectEmail"`
	ProjectHead  []BasicProjectHeadInfo `json:"projectHead"`
}

type UpdateProjectContactInfoResponse struct {
	Data UpdateProjectContactInfo `json:"data"`
}

func ToUpdateProjectContactInfo(project *model.Project) UpdateProjectContactInfo {
	projectHeads := make([]BasicProjectHeadInfo, 0, len(project.Heads))
	for _, v := range project.Heads {
		projectHeads = append(projectHeads, BasicProjectHeadInfo{
			EmployeeID:  v.Employee.ID.String(),
			FullName:    v.Employee.FullName,
			Avatar:      v.Employee.Avatar,
			DisplayName: v.Employee.DisplayName,
			Position:    v.Position,
			Username:    v.Employee.Username,
		})
	}

	var clientEmail []string
	if project.ClientEmail != "" {
		clientEmail = strings.Split(project.ClientEmail, ",")
	}

	return UpdateProjectContactInfo{
		ClientEmail:  clientEmail,
		ProjectEmail: project.ProjectEmail,
		ProjectHead:  projectHeads,
	}
}

type BasicProjectInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Code   string `json:"code"`
	Avatar string `json:"avatar"`
}

func toBasicProjectInfo(project model.Project) *BasicProjectInfo {
	return &BasicProjectInfo{
		ID:     project.ID.String(),
		Type:   project.Type.String(),
		Name:   project.Name,
		Status: project.Status.String(),
		Code:   project.Code,
		Avatar: project.Avatar,
	}
}

type ProjectContentData struct {
	Url string `json:"url"`
}

type ProjectContentDataResponse struct {
	Data *ProjectContentData `json:"data"`
}

func ToProjectContentData(url string) *ProjectContentData {
	return &ProjectContentData{
		Url: url,
	}
}
