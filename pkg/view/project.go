package view

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
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
	AccountManagers     []ProjectHead         `json:"accountManagers"`
	DeliveryManagers    []ProjectHead         `json:"deliveryManagers"`
	SalePersons         []ProjectHead         `json:"salePersons"`
	Stacks              []Stack               `json:"stacks"`
	Code                string                `json:"code"`
	Function            string                `json:"function"`
	AuditNotionID       string                `json:"auditNotionID"`
	BankAccount         *BasicBankAccountInfo `json:"bankAccount"`
	Client              *BasicClientInfo      `json:"client"`
	CompanyInfo         *BasicCompanyInfo     `json:"companyInfo"`
	Organization        *Organization         `json:"organization"`
}

type BasicClientInfo struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	RegistrationNumber string `json:"registrationNumber"`
}

func ToBasicClientInfo(client *model.Client) *BasicClientInfo {
	return &BasicClientInfo{
		ID:                 client.ID.String(),
		Name:               client.Name,
		Description:        client.Description,
		RegistrationNumber: client.RegistrationNumber,
	}
}

type BasicCompanyInfo struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	RegistrationNumber string `json:"registrationNumber"`
}

func ToBasicCompanyInfo(company *model.CompanyInfo) *BasicCompanyInfo {
	return &BasicCompanyInfo{
		ID:                 company.ID.String(),
		Name:               company.Name,
		Description:        company.Description,
		RegistrationNumber: company.RegistrationNumber,
	}
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
	ProjectMemberID      string          `json:"projectMemberID"`
	ProjectSlotID        string          `json:"projectSlotID"`
	EmployeeID           string          `json:"employeeID"`
	FullName             string          `json:"fullName"`
	DisplayName          string          `json:"displayName"`
	Avatar               string          `json:"avatar"`
	Username             string          `json:"username"`
	Status               string          `json:"status"`
	IsLead               bool            `json:"isLead"`
	DeploymentType       string          `json:"deploymentType"`
	StartDate            *time.Time      `json:"startDate"`
	EndDate              *time.Time      `json:"endDate"`
	Rate                 decimal.Decimal `json:"rate"`
	Discount             decimal.Decimal `json:"discount"`
	UpsellCommissionRate decimal.Decimal `json:"upsellCommissionRate"`
	LeadCommissionRate   decimal.Decimal `json:"leadCommissionRate"`
	Currency             *Currency       `json:"currency"`
	Note                 string          `json:"note"`

	Seniority    *model.Seniority   `json:"seniority"`
	Positions    []Position         `json:"positions"`
	UpsellPerson *BasicEmployeeInfo `json:"upsellPerson"`
}

type ProjectHead struct {
	EmployeeID     string          `json:"employeeID"`
	FullName       string          `json:"fullName"`
	DisplayName    string          `json:"displayName"`
	Avatar         string          `json:"avatar"`
	Username       string          `json:"username"`
	CommissionRate decimal.Decimal `json:"commissionRate"`
}

func ToProjectHead(userInfo *model.CurrentLoggedUserInfo, head *model.ProjectHead) ProjectHead {
	res := ProjectHead{
		EmployeeID:  head.EmployeeID.String(),
		FullName:    head.Employee.FullName,
		DisplayName: head.Employee.DisplayName,
		Avatar:      head.Employee.Avatar,
		Username:    head.Employee.Username,
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateRead) {
		res.CommissionRate = head.CommissionRate
	}

	return res
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

func ToProjectData(project *model.Project, userInfo *model.CurrentLoggedUserInfo) ProjectData {
	leadMap := map[string]bool{}
	var technicalLeads = make([]ProjectHead, 0, len(project.Heads))
	var accountManagers, salePersons, deliveryManagers []ProjectHead

	for _, h := range project.Heads {
		head := ToProjectHead(userInfo, h)

		switch h.Position {
		case model.HeadPositionTechnicalLead:
			leadMap[h.EmployeeID.String()] = true
			technicalLeads = append(technicalLeads, head)
		case model.HeadPositionAccountManager:
			accountManagers = append(accountManagers, head)
		case model.HeadPositionDeliveryManager:
			deliveryManagers = append(deliveryManagers, head)
		case model.HeadPositionSalePerson:
			salePersons = append(salePersons, head)
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

		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadFullAccess) {
			member.DeploymentType = m.DeploymentType.String()
		}

		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateRead) && m.UpsellPerson != nil {
			member.UpsellPerson = toBasicEmployeeInfo(*m.UpsellPerson)
		}

		members = append(members, member)
	}

	d := ProjectData{
		BaseModel:        project.BaseModel,
		Avatar:           project.Avatar,
		Name:             project.Name,
		Type:             project.Type.String(),
		Status:           project.Status.String(),
		Stacks:           ToProjectStacks(project.ProjectStacks),
		StartDate:        project.StartDate,
		EndDate:          project.EndDate,
		Members:          members,
		TechnicalLead:    technicalLeads,
		DeliveryManagers: deliveryManagers,
		AccountManagers:  accountManagers,
		SalePersons:      salePersons,
		ProjectEmail:     project.ProjectEmail,

		AllowsSendingSurvey: project.AllowsSendingSurvey,
		Code:                project.Code,
		Function:            project.Function.String(),
	}

	var clientEmail []string
	if project.ClientEmail != "" {
		clientEmail = strings.Split(project.ClientEmail, ",")
	}

	if project.Organization != nil {
		d.Organization = &Organization{
			ID:     project.Organization.ID.String(),
			Code:   project.Organization.Code,
			Name:   project.Organization.Name,
			Avatar: project.Organization.Avatar,
		}
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadFullAccess) {
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
			d.Client = ToBasicClientInfo(project.Client)
		}

		if project.CompanyInfo != nil {
			d.CompanyInfo = ToBasicCompanyInfo(project.CompanyInfo)
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

func ToProjectsData(projects []*model.Project, userInfo *model.CurrentLoggedUserInfo) []ProjectData {
	var results = make([]ProjectData, 0, len(projects))

	for _, p := range projects {
		// If the project belongs user, append it in the list
		_, ok := userInfo.Projects[p.ID]
		if ok && p.Status == model.ProjectStatusActive && model.IsUserActiveInProject(userInfo.UserID, p.ProjectMembers) {
			results = append(results, ToProjectData(p, userInfo))
			continue
		}

		// If the project is not belong user, check if the user has permission to view the project
		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadFullAccess) ||
			(authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadReadActive) &&
				p.Status == model.ProjectStatusActive) {
			results = append(results, ToProjectData(p, userInfo))
			continue
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
	ProjectSlotID        string             `json:"projectSlotID"`
	ProjectMemberID      string             `json:"projectMemberID"`
	EmployeeID           string             `json:"employeeID"`
	FullName             string             `json:"fullName"`
	DisplayName          string             `json:"displayName"`
	Avatar               string             `json:"avatar"`
	Positions            []Position         `json:"positions"`
	DeploymentType       string             `json:"deploymentType"`
	Status               string             `json:"status"`
	IsLead               bool               `json:"isLead"`
	Seniority            model.Seniority    `json:"seniority"`
	Username             string             `json:"username"`
	Rate                 decimal.Decimal    `json:"rate"`
	Discount             decimal.Decimal    `json:"discount"`
	UpsellPerson         *BasicEmployeeInfo `json:"upsellPerson"`
	UpsellCommissionRate decimal.Decimal    `json:"upsellCommissionRate"`
	LeadCommissionRate   decimal.Decimal    `json:"leadCommissionRate"`
	Note                 string             `json:"note"`
}

type CreateMemberDataResponse struct {
	Data CreateMemberData `json:"data"`
}

func ToCreateMemberData(userInfo *model.CurrentLoggedUserInfo, slot *model.ProjectSlot) CreateMemberData {
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
		Note:           slot.Note,
	}

	if !slot.ProjectMember.ID.IsZero() {
		rs.ProjectMemberID = slot.ProjectMember.ID.String()
		rs.EmployeeID = slot.ProjectMember.EmployeeID.String()
		rs.Note = slot.ProjectMember.Note
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateRead) &&
		slot.ProjectMember.UpsellPerson != nil {
		rs.UpsellPerson = toBasicEmployeeInfo(*slot.ProjectMember.UpsellPerson)
		rs.UpsellCommissionRate = slot.ProjectMember.UpsellCommissionRate
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectMembersRateRead) {
		rs.Rate = slot.Rate
		rs.Discount = slot.Discount

		if !slot.ProjectMember.ID.IsZero() {
			rs.Rate = slot.ProjectMember.Rate
			rs.Discount = slot.ProjectMember.Discount
		}
	}

	return rs
}

type CreateProjectData struct {
	model.BaseModel

	Name             string                `json:"name"`
	Type             string                `json:"type"`
	Status           string                `json:"status"`
	StartDate        string                `json:"startDate"`
	AccountManagers  []ProjectHead         `json:"accountManagers"`
	DeliveryManagers []ProjectHead         `json:"deliveryManagers"`
	SalePersons      []ProjectHead         `json:"salePersons"`
	Members          []CreateMemberData    `json:"members"`
	ClientEmail      []string              `json:"clientEmail"`
	ProjectEmail     string                `json:"projectEmail"`
	Country          *BasicCountryInfo     `json:"country"`
	Code             string                `json:"code"`
	Function         string                `json:"function"`
	BankAccount      *BasicBankAccountInfo `json:"bankAccount"`
	Client           *Client               `json:"client"`
	Organization     *Organization         `json:"organization"`
}

type BasicBankAccountInfo struct {
	ID            string `json:"id"`
	AccountNumber string `json:"accountNumber"`
	BankName      string `json:"bankName"`
	OwnerName     string `json:"ownerName"`
}

func ToCreateProjectDataResponse(userInfo *model.CurrentLoggedUserInfo, project *model.Project) CreateProjectData {
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

	if project.Organization != nil {
		result.Organization = &Organization{
			ID:     project.Organization.ID.String(),
			Code:   project.Organization.Code,
			Name:   project.Organization.Name,
			Avatar: project.Organization.Avatar,
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
			result.AccountManagers = append(result.AccountManagers, ToProjectHead(userInfo, head))
		case model.HeadPositionDeliveryManager:
			result.DeliveryManagers = append(result.DeliveryManagers, ToProjectHead(userInfo, head))
		case model.HeadPositionSalePerson:
			result.SalePersons = append(result.SalePersons, ToProjectHead(userInfo, head))
		}
	}

	result.Members = make([]CreateMemberData, 0, len(project.Slots))
	for _, slot := range project.Slots {
		result.Members = append(result.Members, ToCreateMemberData(userInfo, &slot))
	}

	return result
}

func ToProjectMemberListData(userInfo *model.CurrentLoggedUserInfo, members []*model.ProjectMember, projectHeads []*model.ProjectHead, project *model.Project, distinct bool) []ProjectMember {
	var results = make([]ProjectMember, 0, len(members))

	leadMap := map[string]*model.ProjectHead{}
	for _, v := range projectHeads {
		if v.IsLead() {
			leadMap[v.EmployeeID.String()] = v
		}
	}

	for _, m := range members {
		var member ProjectMember

		if m.ID.IsZero() {
			member = ProjectMember{
				ProjectSlotID:  m.ProjectSlotID.String(),
				Status:         m.Status.String(),
				DeploymentType: m.DeploymentType.String(),
				Seniority:      m.Seniority,
				Note:           m.Note,
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
				IsLead:          leadMap[m.EmployeeID.String()] != nil,
				Status:          m.Status.String(),
				DeploymentType:  m.DeploymentType.String(),
				Seniority:       m.Seniority,
				Note:            m.Note,
				Positions:       ToProjectMemberPositions(m.ProjectMemberPositions),
			}
		}

		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadFullAccess) &&
			project.BankAccount != nil &&
			project.BankAccount.Currency != nil {
			member.Currency = new(Currency)
			*member.Currency = toCurrency(project.BankAccount.Currency)
		}

		// add commission rate
		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateRead) {
			if leadMap[m.EmployeeID.String()] != nil {
				member.LeadCommissionRate = leadMap[m.EmployeeID.String()].CommissionRate
			}

			member.UpsellCommissionRate = m.UpsellCommissionRate
			if m.UpsellPerson != nil {
				member.UpsellPerson = toBasicEmployeeInfo(*m.UpsellPerson)
			}
		}

		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectMembersRateRead) {
			member.Rate = m.Rate
			member.Discount = m.Discount
		}

		results = append(results, member)
	}

	// Remove duplicate members
	if distinct {
		uniqueResults := make([]ProjectMember, 0, len(results))
		uniqueMap := map[string]bool{}
		for _, v := range results {
			if _, ok := uniqueMap[v.EmployeeID]; !ok {
				uniqueMap[v.EmployeeID] = true
				uniqueResults = append(uniqueResults, v)
			}
		}

		return uniqueResults
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
	Organization  *Organization         `json:"organization"`
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

	if project.Organization != nil {
		rs.Organization = &Organization{
			ID:     project.Organization.ID.String(),
			Code:   project.Organization.Code,
			Name:   project.Organization.Name,
			Avatar: project.Organization.Avatar,
		}
	}

	if project.Client != nil {
		rs.Client = toClient(project.Client)
	}

	return rs
}

type BasicProjectHeadInfo struct {
	EmployeeID     string             `json:"employeeID"`
	FullName       string             `json:"fullName"`
	DisplayName    string             `json:"displayName"`
	Avatar         string             `json:"avatar"`
	Position       model.HeadPosition `json:"position"`
	Username       string             `json:"username"`
	CommissionRate decimal.Decimal    `json:"commissionRate"`
}

type UpdateProjectContactInfo struct {
	ClientEmail  []string               `json:"clientEmail"`
	ProjectEmail string                 `json:"projectEmail"`
	ProjectHead  []BasicProjectHeadInfo `json:"projectHead"`
}

type UpdateProjectContactInfoResponse struct {
	Data UpdateProjectContactInfo `json:"data"`
}

func ToUpdateProjectContactInfo(project *model.Project, userInfo *model.CurrentLoggedUserInfo) UpdateProjectContactInfo {
	projectHeads := make([]BasicProjectHeadInfo, 0, len(project.Heads))
	for _, v := range project.Heads {
		ph := BasicProjectHeadInfo{
			EmployeeID:  v.Employee.ID.String(),
			FullName:    v.Employee.FullName,
			Avatar:      v.Employee.Avatar,
			DisplayName: v.Employee.DisplayName,
			Position:    v.Position,
			Username:    v.Employee.Username,
		}

		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateRead) {
			ph.CommissionRate = v.CommissionRate
		}

		projectHeads = append(projectHeads, ph)
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
