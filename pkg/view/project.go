package view

import (
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/gin-gonic/gin"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/shopspring/decimal"
)

type ProjectData struct {
	model.BaseModel

	Name                string            `json:"name"`
	Avatar              string            `json:"avatar"`
	Type                string            `json:"type"`
	Status              string            `json:"status"`
	ProjectEmail        string            `json:"projectEmail"`
	ClientEmail         []string          `json:"clientEmail"`
	Industry            string            `json:"industry"`
	AllowsSendingSurvey bool              `json:"allowsSendingSurvey"`
	Country             *BasicCountryInfo `json:"country"`
	StartDate           *time.Time        `json:"startDate"`
	EndDate             *time.Time        `json:"endDate"`
	Members             []ProjectMember   `json:"members"`
	TechnicalLead       []ProjectHead     `json:"technicalLeads"`
	AccountManager      *ProjectHead      `json:"accountManager"`
	SalePerson          *ProjectHead      `json:"salePerson"`
	DeliveryManager     *ProjectHead      `json:"deliveryManager"`
	Stacks              []Stack           `json:"stacks"`
	Code                string            `json:"code"`
	Function            string            `json:"function"`
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
	JoinedDate      *time.Time      `json:"joinedDate"`
	LeftDate        *time.Time      `json:"leftDate"`
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

		if h.Employee.WorkingStatus == model.WorkingStatusLeft {
			continue
		}

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

	var members = make([]ProjectMember, 0, len(project.Slots))
	for _, slot := range project.Slots {
		m := slot.ProjectMember
		member := ProjectMember{
			Status:         slot.Status.String(),
			DeploymentType: slot.DeploymentType.String(),
			Positions:      ToProjectSlotPositions(slot.ProjectSlotPositions),
		}

		if !slot.Seniority.ID.IsZero() {
			member.Seniority = &slot.Seniority
		}

		if slot.Status != model.ProjectMemberStatusPending && !m.ID.IsZero() {
			member.DeploymentType = ""

			if utils.HasPermission(c, userInfo.Permissions, "projectWorkUnits.read.fullAccess") {
				member.DeploymentType = m.DeploymentType.String()
			}

			member.Status = m.Status.String()
			member.EmployeeID = m.EmployeeID.String()
			member.FullName = m.Employee.FullName
			member.DisplayName = m.Employee.DisplayName
			member.Avatar = m.Employee.Avatar
			member.Username = m.Employee.Username
			member.Seniority = m.Seniority
			member.IsLead = leadMap[m.EmployeeID.String()]
			member.Positions = ToProjectMemberPositions(m.ProjectMemberPositions)
		}

		if m.Status == model.ProjectMemberStatusInactive {
			member.LeftDate = m.LeftDate
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

	if utils.HasPermission(c, userInfo.Permissions, "projectWorkUnits.read.fullAccess") {
		d.ClientEmail = clientEmail
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
		if ok && p.Status == model.ProjectStatusActive && model.IsUserActiveInProject(userInfo.UserID, p.Slots) {
			results = append(results, ToProjectData(c, p, userInfo))
			continue
		}

		// If the project is not belong user, check if the user has permission to view the project
		if utils.HasPermission(c, userInfo.Permissions, "projects.read.fullAccess") {
			results = append(results, ToProjectData(c, p, userInfo))
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

	Name            string             `json:"name"`
	Type            string             `json:"type"`
	Status          string             `json:"status"`
	StartDate       string             `json:"startDate"`
	AccountManager  *ProjectHead       `json:"accountManager"`
	DeliveryManager *ProjectHead       `json:"deliveryManager"`
	Members         []CreateMemberData `json:"members"`
	ClientEmail     []string           `json:"clientEmail"`
	ProjectEmail    string             `json:"projectEmail"`
	Country         *BasicCountryInfo  `json:"country"`
	Code            string             `json:"code"`
	Function        string             `json:"function"`
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

func ToProjectMemberListData(slots []*model.ProjectSlot, projectHeads []*model.ProjectHead) []ProjectMember {
	var results = make([]ProjectMember, 0, len(slots))

	leadMap := map[string]bool{}
	for _, v := range projectHeads {
		if v.IsLead() {
			leadMap[v.EmployeeID.String()] = true
		}
	}

	for _, slot := range slots {
		m := slot.ProjectMember

		member := ProjectMember{
			ProjectSlotID:  slot.ID.String(),
			Status:         slot.Status.String(),
			DeploymentType: slot.DeploymentType.String(),
			Rate:           slot.Rate,
			Discount:       slot.Discount,
			Positions:      ToProjectSlotPositions(slot.ProjectSlotPositions),
		}

		if !slot.Seniority.ID.IsZero() {
			member.Seniority = &slot.Seniority
		}

		if slot.Status != model.ProjectMemberStatusPending && !m.ID.IsZero() {
			member.ProjectMemberID = m.ID.String()
			member.EmployeeID = m.EmployeeID.String()
			member.FullName = m.Employee.FullName
			member.DisplayName = m.Employee.DisplayName
			member.Avatar = m.Employee.Avatar
			member.Username = m.Employee.Username
			member.JoinedDate = m.JoinedDate
			member.IsLead = leadMap[m.EmployeeID.String()]
			member.Rate = m.Rate
			member.Discount = m.Discount
			member.Status = m.Status.String()
			member.Positions = ToProjectMemberPositions(m.ProjectMemberPositions)
		}

		if m.Status == model.ProjectMemberStatusInactive {
			member.LeftDate = m.LeftDate
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
	Name      string                `json:"name"`
	StartDate *time.Time            `json:"startDate"`
	Country   *BasicCountryInfo     `json:"country"`
	Stacks    []model.Stack         `json:"stacks"`
	Function  model.ProjectFunction `json:"function"`
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

	if project.Country != nil {
		rs.Country = &BasicCountryInfo{
			ID:   project.Country.ID,
			Name: project.Country.Name,
			Code: project.Country.Code,
		}
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
}

func toBasicProjectInfo(project model.Project) *BasicProjectInfo {
	return &BasicProjectInfo{
		ID:     project.ID.String(),
		Type:   project.Type.String(),
		Name:   project.Name,
		Status: project.Status.String(),
		Code:   project.Code,
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
