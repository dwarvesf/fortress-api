package view

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
)

type Project struct {
	ID                  string                   `json:"id"`
	CreatedAt           time.Time                `json:"createdAt"`
	UpdatedAt           *time.Time               `json:"updatedAt"`
	Name                string                   `json:"name"`
	CountryID           string                   `json:"countryID"`
	Type                string                   `json:"type"`
	StartDate           *time.Time               `json:"startDate"`
	EndDate             *time.Time               `json:"end_date"`
	Status              string                   `json:"status"`
	ProjectEmail        string                   `json:"projectEmail"`
	ClientEmail         string                   `json:"clientEmail"`
	Avatar              string                   `json:"avatar"`
	AllowsSendingSurvey bool                     `json:"allowsSendingSurvey"`
	Code                string                   `json:"code"`
	Function            string                   `json:"function"`
	BankAccountID       string                   `json:"bankAccountID"`
	CompanyInfoID       string                   `json:"companyInfoID"`
	ClientID            string                   `json:"clientID"`
	OrganizationID      string                   `json:"organizationID"`
	AccountRating       int                      `json:"accountRating"`
	DeliveryRating      int                      `json:"deliveryRating"`
	LeadRating          int                      `json:"leadRating"`
	ImportantLevel      string                   `json:"importantLevel"`
	ProjectNotion       *ProjectNotion           `json:"projectNotion"`
	Organization        *Organization            `json:"organization"`
	BankAccount         *BankAccount             `json:"bankAccount"`
	Country             *Country                 `json:"country"`
	Client              *Client                  `json:"client"`
	CompanyInfo         *CompanyInfo             `json:"companyInfo"`
	Slots               []ProjectSlot            `json:"slots"`
	Heads               []*ProjectHead           `json:"heads"`
	ProjectMembers      []ProjectMember          `json:"projectMembers"`
	ProjectStacks       []Stack                  `json:"projectStacks"`
	CommissionConfigs   ProjectCommissionConfigs `json:"commissionConfigs"`
	ProjectInfo         *ProjectInfo             `json:"projectInfo"`
} // @name Project

func ToProjects(projects []model.Project) []Project {
	rs := make([]Project, 0, len(projects))
	for _, project := range projects {
		rs = append(rs, *ToProject(&project))
	}

	return rs
}

func ToProject(project *model.Project) *Project {
	if project == nil {
		return nil
	}

	return &Project{
		ID:                  project.ID.String(),
		CreatedAt:           project.CreatedAt,
		UpdatedAt:           project.UpdatedAt,
		Name:                project.Name,
		CountryID:           project.CountryID.String(),
		Type:                project.Type.String(),
		StartDate:           project.StartDate,
		EndDate:             project.EndDate,
		Status:              project.Status.String(),
		ProjectEmail:        project.ProjectEmail,
		ClientEmail:         project.ClientEmail,
		Avatar:              project.Avatar,
		AllowsSendingSurvey: project.AllowsSendingSurvey,
		Code:                project.Code,
		Function:            project.Function.String(),
		BankAccountID:       project.BankAccountID.String(),
		CompanyInfoID:       project.CompanyInfoID.String(),
		ClientID:            project.ClientID.String(),
		OrganizationID:      project.OrganizationID.String(),
		AccountRating:       project.AccountRating,
		DeliveryRating:      project.DeliveryRating,
		LeadRating:          project.LeadRating,
		ImportantLevel:      project.ImportantLevel.String(),
		ProjectNotion:       ToProjectNotion(project.ProjectNotion),
		Organization:        ToOrganization(project.Organization),
		BankAccount:         ToBankAccount(project.BankAccount),
		Country:             ToCountry(project.Country),
		Client:              ToClient(project.Client),
		CompanyInfo:         ToCompanyInfo(project.CompanyInfo),
		Slots:               ToProjectSlotList(project.Slots),
		Heads:               ToProjectHeads(project.Heads),
		ProjectMembers:      ToProjectMembers(project.ProjectMembers),
		ProjectStacks:       ToProjectStacks(project.ProjectStacks),
		CommissionConfigs:   ToProjectCommissionConfigs(project.CommissionConfigs),
		ProjectInfo:         ToProjectInfo(project.ProjectInfo),
	}
}

type ProjectInfo struct {
	ID                     string     `json:"id"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              *time.Time `json:"updatedAt"`
	ProjectID              *string    `json:"projectID"`
	BasecampBucketID       int64      `json:"basecampBucketID"`
	BasecampScheduleID     int64      `json:"basecampScheduleID"`
	BasecampCampfireID     int64      `json:"basecampCampfireID"`
	BasecampTodolistID     int64      `json:"basecampTodolistID"`
	BasecampMessageBoardID int64      `json:"basecampMessageBoardID"`
	BasecampSentryID       int64      `json:"basecampSentryID"`
	GitlabID               int64      `json:"gitlabID"`
	Repositories           []byte     `json:"repositories"`
	Project                *Project   `json:"project"`
} // @name ProjectInfo

func ToProjectInfo(projectInfo *model.ProjectInfo) *ProjectInfo {
	if projectInfo == nil {
		return nil
	}

	var projectID *string
	if projectInfo.ProjectID != nil {
		id := projectInfo.ProjectID.String()
		projectID = &id
	}

	return &ProjectInfo{
		ID:                     projectInfo.ID.String(),
		CreatedAt:              projectInfo.CreatedAt,
		UpdatedAt:              projectInfo.UpdatedAt,
		ProjectID:              projectID,
		BasecampBucketID:       projectInfo.BasecampBucketID,
		BasecampScheduleID:     projectInfo.BasecampScheduleID,
		BasecampCampfireID:     projectInfo.BasecampCampfireID,
		BasecampTodolistID:     projectInfo.BasecampTodolistID,
		BasecampMessageBoardID: projectInfo.BasecampMessageBoardID,
		BasecampSentryID:       projectInfo.BasecampSentryID,
		GitlabID:               projectInfo.GitlabID,
		Repositories:           projectInfo.Repositories,
		Project:                ToProject(projectInfo.Project),
	}
}

type ProjectCommissionConfigs []ProjectCommissionConfig // @name ProjectCommissionConfigs

func ToProjectCommissionConfigs(configs []model.ProjectCommissionConfig) ProjectCommissionConfigs {
	rs := make(ProjectCommissionConfigs, 0, len(configs))
	for _, config := range configs {
		rs = append(rs, ToProjectCommissionConfig(&config))
	}

	return rs
}

type ProjectCommissionConfig struct {
	ID             string          `json:"id"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      *time.Time      `json:"updatedAt"`
	ProjectID      string          `json:"projectID"`
	Position       string          `json:"position"`
	CommissionRate decimal.Decimal `json:"commissionRate"`
} // @name ProjectCommissionConfig

func ToProjectCommissionConfig(config *model.ProjectCommissionConfig) ProjectCommissionConfig {
	if config == nil {
		return ProjectCommissionConfig{}
	}

	return ProjectCommissionConfig{
		ID:             config.ID.String(),
		CreatedAt:      config.CreatedAt,
		UpdatedAt:      config.UpdatedAt,
		ProjectID:      config.ProjectID.String(),
		Position:       config.Position.String(),
		CommissionRate: config.CommissionRate,
	}
}

type ProjectNotion struct {
	ID            string     `json:"id"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     *time.Time `json:"updatedAt"`
	ProjectID     string     `json:"projectID"`
	AuditNotionID string     `json:"auditNotionID"`
	Project       *Project   `json:"project"`
} // @name ProjectNotion

func ToProjectNotion(p *model.ProjectNotion) *ProjectNotion {
	if p == nil {
		return nil
	}

	return &ProjectNotion{
		ID:            p.ID.String(),
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		ProjectID:     p.ProjectID.String(),
		AuditNotionID: p.AuditNotionID.String(),
		Project:       ToProject(p.Project),
	}
}

type ProjectSlot struct {
	ID                   string          `json:"id"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            *time.Time      `json:"updatedAt"`
	ProjectID            string          `json:"projectID"`
	SeniorityID          string          `json:"seniorityID"`
	UpsellPersonID       string          `json:"upsellPersonID"`
	DeploymentType       string          `json:"deploymentType"`
	Status               string          `json:"status"`
	Rate                 decimal.Decimal `json:"rate"`
	Discount             decimal.Decimal `json:"discount"`
	Note                 string          `json:"note"`
	Seniority            Seniority       `json:"seniority"`
	Project              Project         `json:"project"`
	ProjectMember        ProjectMember   `json:"projectMember"`
	ProjectSlotPositions []Position      `json:"projectSlotPositions"`
	UpsellPerson         *EmployeeData   `json:"upsellPerson"`
} // @name ProjectSlot

func ToProjectSlot(slot *model.ProjectSlot) *ProjectSlot {
	if slot == nil {
		return nil
	}

	project := ToProject(&slot.Project)
	if project == nil {
		// make sure project is not nil
		project = &Project{}
	}

	projectMember := ToProjectMember(&slot.ProjectMember)
	if projectMember == nil {
		// make sure project member is not nil
		projectMember = &ProjectMember{}
	}

	return &ProjectSlot{
		ID:                   slot.ID.String(),
		CreatedAt:            slot.CreatedAt,
		UpdatedAt:            slot.UpdatedAt,
		ProjectID:            slot.ProjectID.String(),
		SeniorityID:          slot.SeniorityID.String(),
		UpsellPersonID:       slot.UpsellPersonID.String(),
		DeploymentType:       slot.DeploymentType.String(),
		Status:               slot.Status.String(),
		Rate:                 slot.Rate,
		Discount:             slot.Discount,
		Note:                 slot.Note,
		Seniority:            ToSeniority(slot.Seniority),
		Project:              *project,
		ProjectMember:        *projectMember,
		ProjectSlotPositions: ToProjectSlotPositions(slot.ProjectSlotPositions),
		UpsellPerson:         ToEmployeeData(slot.UpsellPerson),
	}
}

func ToProjectSlotList(slots []model.ProjectSlot) []ProjectSlot {
	rs := make([]ProjectSlot, 0, len(slots))
	for _, slot := range slots {
		projectSlot := ToProjectSlot(&slot)
		if projectSlot != nil {
			rs = append(rs, *projectSlot)
		}
	}

	return rs
}

type ProjectData struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`

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
	AccountRating       int                   `json:"accountRating"`
	DeliveryRating      int                   `json:"deliveryRating"`
	LeadRating          int                   `json:"leadRating"`
	ImportantLevel      string                `json:"importantLevel"`
	BankAccount         *BasicBankAccountInfo `json:"bankAccount"`
	Client              *BasicClientInfo      `json:"client"`
	CompanyInfo         *BasicCompanyInfo     `json:"companyInfo"`
	Organization        *Organization         `json:"organization"`
	MonthlyChargeRate   decimal.Decimal       `json:"monthlyChargeRate"`
	Currency            *Currency             `json:"currency"`
} // @name ProjectData

type BasicClientInfo struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	RegistrationNumber string          `json:"registrationNumber"`
	Address            string          `json:"address"`
	Contacts           []ClientContact `json:"contacts"`
} // @name BasicClientInfo

func ToBasicClientInfo(client *model.Client) *BasicClientInfo {
	if client == nil {
		return nil
	}

	clientContacts := make([]ClientContact, 0, len(client.Contacts))
	for _, contact := range client.Contacts {
		emails := make([]string, 0)
		_ = json.Unmarshal(contact.Emails, &emails)

		clientContacts = append(clientContacts, ClientContact{
			ID:     contact.ID.String(),
			Name:   contact.Name,
			Role:   contact.Role,
			Emails: emails,
		})
	}

	return &BasicClientInfo{
		ID:                 client.ID.String(),
		Name:               client.Name,
		Description:        client.Description,
		RegistrationNumber: client.RegistrationNumber,
		Address:            client.Address,
		Contacts:           clientContacts,
	}
}

type BasicCompanyInfo struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	RegistrationNumber string `json:"registrationNumber"`
} // @name BasicCompanyInfo

func ToBasicCompanyInfo(company *model.CompanyInfo) *BasicCompanyInfo {
	return &BasicCompanyInfo{
		ID:                 company.ID.String(),
		Name:               company.Name,
		Description:        company.Description,
		RegistrationNumber: company.RegistrationNumber,
	}
}

type UpdatedProject struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`

	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
} // @name UpdatedProject

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

	Seniority    *Seniority         `json:"seniority"`
	Positions    []Position         `json:"positions"`
	UpsellPerson *BasicEmployeeInfo `json:"upsellPerson"`
} // @name ProjectMember

func ToProjectMembers(members []model.ProjectMember) []ProjectMember {
	res := make([]ProjectMember, 0, len(members))
	for _, m := range members {
		res = append(res, *ToProjectMember(&m))
	}

	return res
}

func ToProjectMember(member *model.ProjectMember) *ProjectMember {
	var seniority *Seniority
	if member.Seniority != nil {
		s := ToSeniority(*member.Seniority)
		seniority = &s
	}

	return &ProjectMember{
		ProjectMemberID:      member.ID.String(),
		ProjectSlotID:        member.ProjectSlotID.String(),
		EmployeeID:           member.EmployeeID.String(),
		FullName:             member.Employee.FullName,
		DisplayName:          member.Employee.DisplayName,
		Avatar:               member.Employee.Avatar,
		Username:             member.Employee.Username,
		Status:               member.Status.String(),
		IsLead:               member.IsLead,
		DeploymentType:       member.DeploymentType.String(),
		StartDate:            member.StartDate,
		EndDate:              member.EndDate,
		Rate:                 member.Rate,
		Discount:             member.Discount,
		UpsellCommissionRate: member.UpsellCommissionRate,
		Note:                 member.Note,
		Seniority:            seniority,
		Positions:            ToProjectMemberPositions(member.ProjectMemberPositions),
		UpsellPerson:         toBasicEmployeeInfo(*member.UpsellPerson),
	}
}

type ProjectHead struct {
	EmployeeID          string          `json:"employeeID"`
	FullName            string          `json:"fullName"`
	DisplayName         string          `json:"displayName"`
	Avatar              string          `json:"avatar"`
	Username            string          `json:"username"`
	CommissionRate      decimal.Decimal `json:"commissionRate"`
	FinalCommissionRate decimal.Decimal `json:"finalCommissionRate"`
} // @name ProjectHead

func ToProjectHeads(heads []*model.ProjectHead) []*ProjectHead {
	res := make([]*ProjectHead, 0, len(heads))
	for _, h := range heads {
		ph := ToProjectHead(nil, h, nil)
		res = append(res, &ph)
	}

	return res
}

func ToProjectHead(userInfo *model.CurrentLoggedUserInfo, head *model.ProjectHead, commissionConfig map[string]decimal.Decimal) ProjectHead {
	res := ProjectHead{
		EmployeeID:  head.EmployeeID.String(),
		FullName:    head.Employee.FullName,
		DisplayName: head.Employee.DisplayName,
		Avatar:      head.Employee.Avatar,
		Username:    head.Employee.Username,
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateRead) {
		res.CommissionRate = head.CommissionRate
		rate := decimal.Zero
		v, ok := commissionConfig[head.Position.String()]
		if ok {
			rate = v.Mul(head.CommissionRate).Div(decimal.NewFromInt(100))
		}
		res.FinalCommissionRate = rate
	}

	return res
}

type UpdateProjectStatusResponse struct {
	Data UpdatedProject `json:"data"`
} // @name UpdateProjectStatusResponse

func ToUpdateProjectStatusResponse(p *model.Project) UpdatedProject {
	return UpdatedProject{
		ID:        p.ID.String(),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,

		Name:      p.Name,
		Type:      p.Type.String(),
		Status:    p.Status.String(),
		StartDate: p.StartDate,
		EndDate:   p.EndDate,
	}
}

func ToProjectData(in *model.Project, userInfo *model.CurrentLoggedUserInfo) ProjectData {
	leadMap := map[string]bool{}
	var technicalLeads = make([]ProjectHead, 0, len(in.Heads))
	var accountManagers, salePersons, deliveryManagers []ProjectHead
	commissionConfig := in.CommissionConfigs.ToMap()
	for _, h := range in.Heads {
		head := ToProjectHead(userInfo, h, commissionConfig)

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

	var projectCurrency *Currency
	if in.BankAccount != nil {
		projectCurrency = toCurrency(in.BankAccount.Currency)
	}

	monthlyRevenue := decimal.Zero
	var members = make([]ProjectMember, 0, len(in.ProjectMembers))
	for _, m := range in.ProjectMembers {
		var seniority *Seniority
		if m.Seniority != nil {
			s := ToSeniority(*m.Seniority)
			seniority = &s
		}

		member := ProjectMember{
			Status:      m.Status.String(),
			EmployeeID:  m.EmployeeID.String(),
			FullName:    m.Employee.FullName,
			DisplayName: m.Employee.DisplayName,
			Avatar:      m.Employee.Avatar,
			Username:    m.Employee.Username,
			Seniority:   seniority,
			IsLead:      leadMap[m.EmployeeID.String()],
			Positions:   ToProjectMemberPositions(m.ProjectMemberPositions),
		}

		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadFullAccess) {
			member.DeploymentType = m.DeploymentType.String()

			if m.UpsellPerson != nil {
				member.UpsellPerson = toBasicEmployeeInfo(*m.UpsellPerson)
			}

			member.Rate = m.Rate
			member.Currency = projectCurrency
		}

		if m.DeploymentType == model.MemberDeploymentTypeOfficial && m.Status == model.ProjectMemberStatusActive {
			monthlyRevenue = monthlyRevenue.Add(m.Rate)
		}

		members = append(members, member)
	}

	projectData := ProjectData{
		ID:                  in.ID.String(),
		CreatedAt:           in.CreatedAt,
		UpdatedAt:           in.UpdatedAt,
		Avatar:              in.Avatar,
		Name:                in.Name,
		Type:                in.Type.String(),
		Status:              in.Status.String(),
		Stacks:              ToProjectStacks(in.ProjectStacks),
		StartDate:           in.StartDate,
		EndDate:             in.EndDate,
		Members:             members,
		TechnicalLead:       technicalLeads,
		DeliveryManagers:    deliveryManagers,
		AccountManagers:     accountManagers,
		SalePersons:         salePersons,
		ProjectEmail:        in.ProjectEmail,
		AllowsSendingSurvey: in.AllowsSendingSurvey,
		Code:                in.Code,
		Function:            in.Function.String(),
		Currency:            projectCurrency,
	}

	var clientEmail []string
	if in.ClientEmail != "" {
		clientEmail = strings.Split(in.ClientEmail, ",")
	}

	if in.Organization != nil {
		projectData.Organization = &Organization{
			ID:     in.Organization.ID.String(),
			Code:   in.Organization.Code,
			Name:   in.Organization.Name,
			Avatar: in.Organization.Avatar,
		}
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadMonthlyRevenue) {
		projectData.MonthlyChargeRate = monthlyRevenue
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadFullAccess) {
		if in.ProjectNotion != nil && !in.ProjectNotion.AuditNotionID.IsZero() {
			projectData.AuditNotionID = in.ProjectNotion.AuditNotionID.String()
		}

		projectData.ClientEmail = clientEmail

		if in.BankAccount != nil {
			projectData.BankAccount = &BasicBankAccountInfo{
				ID:            in.BankAccount.ID.String(),
				AccountNumber: in.BankAccount.AccountNumber,
				BankName:      in.BankAccount.BankName,
				OwnerName:     in.BankAccount.OwnerName,
			}
		}

		if in.Client != nil {
			projectData.Client = ToBasicClientInfo(in.Client)
		}

		if in.CompanyInfo != nil {
			projectData.CompanyInfo = ToBasicCompanyInfo(in.CompanyInfo)
		}

		projectData.AccountRating = in.AccountRating
		projectData.DeliveryRating = in.DeliveryRating
		projectData.LeadRating = in.LeadRating
		projectData.ImportantLevel = in.ImportantLevel.String()
	}

	if in.Country != nil {
		projectData.Country = &BasicCountryInfo{
			ID:   UUID(in.Country.ID),
			Name: in.Country.Name,
			Code: in.Country.Code,
		}
	}

	return projectData
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
	PaginationResponse
	Data []ProjectData `json:"data"`
} // @name ProjectListDataResponse

type ProjectDataResponse struct {
	Data ProjectData `json:"data"`
} // @name ProjectDataResponse

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
	Seniority            Seniority          `json:"seniority"`
	Username             string             `json:"username"`
	Rate                 decimal.Decimal    `json:"rate"`
	Discount             decimal.Decimal    `json:"discount"`
	UpsellPerson         *BasicEmployeeInfo `json:"upsellPerson"`
	UpsellCommissionRate decimal.Decimal    `json:"upsellCommissionRate"`
	LeadCommissionRate   decimal.Decimal    `json:"leadCommissionRate"`
	Note                 string             `json:"note"`
} // @name CreateMemberData

type CreateMemberDataResponse struct {
	Data CreateMemberData `json:"data"`
} // @name CreateMemberDataResponse

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
		Seniority:      ToSeniority(slot.Seniority),
		Note:           slot.Note,
	}

	if !slot.ProjectMember.ID.IsZero() {
		rs.ProjectMemberID = slot.ProjectMember.ID.String()
		rs.EmployeeID = slot.ProjectMember.EmployeeID.String()
		rs.Note = slot.ProjectMember.Note

		if slot.ProjectMember.UpsellPerson != nil {
			rs.UpsellPerson = toBasicEmployeeInfo(*slot.ProjectMember.UpsellPerson)
		}
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateRead) {
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

type CreateProjectRestponse struct {
	ID        UUID       `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`

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
} // @name CreateProjectRestponse

type BasicBankAccountInfo struct {
	ID            string `json:"id"`
	AccountNumber string `json:"accountNumber"`
	BankName      string `json:"bankName"`
	OwnerName     string `json:"ownerName"`
} // @name BasicBankAccountInfo

func ToCreateProjectDataResponse(userInfo *model.CurrentLoggedUserInfo, project *model.Project) CreateProjectRestponse {
	var clientEmail []string
	if project.ClientEmail != "" {
		clientEmail = strings.Split(project.ClientEmail, ",")
	}

	result := CreateProjectRestponse{
		ID:           UUID(project.ID),
		CreatedAt:    project.CreatedAt,
		UpdatedAt:    project.UpdatedAt,
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
			ID:   UUID(project.Country.ID),
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
		result.Client = ToClient(project.Client)
	}

	if project.StartDate != nil {
		result.StartDate = project.StartDate.Format("2006-01-02")
	}

	commisionRate := project.CommissionConfigs.ToMap()

	for _, head := range project.Heads {
		switch head.Position {
		case model.HeadPositionAccountManager:
			result.AccountManagers = append(result.AccountManagers, ToProjectHead(userInfo, head, commisionRate))
		case model.HeadPositionDeliveryManager:
			result.DeliveryManagers = append(result.DeliveryManagers, ToProjectHead(userInfo, head, commisionRate))
		case model.HeadPositionSalePerson:
			result.SalePersons = append(result.SalePersons, ToProjectHead(userInfo, head, commisionRate))
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
		var seniority *Seniority
		if m.Seniority != nil {
			s := ToSeniority(*m.Seniority)
			seniority = &s
		}

		if m.ID.IsZero() {
			member = ProjectMember{
				ProjectSlotID:  m.ProjectSlotID.String(),
				Status:         m.Status.String(),
				DeploymentType: m.DeploymentType.String(),
				Seniority:      seniority,
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
				Seniority:       seniority,
				Note:            m.Note,
				Positions:       ToProjectMemberPositions(m.ProjectMemberPositions),
			}
		}

		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadFullAccess) &&
			project.BankAccount != nil &&
			project.BankAccount.Currency != nil {
			member.Currency = toCurrency(project.BankAccount.Currency)

			if m.UpsellPerson != nil {
				member.UpsellPerson = toBasicEmployeeInfo(*m.UpsellPerson)
			}
		}

		// add commission rate
		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateRead) {
			if leadMap[m.EmployeeID.String()] != nil {
				member.LeadCommissionRate = leadMap[m.EmployeeID.String()].CommissionRate
			}

			member.UpsellCommissionRate = m.UpsellCommissionRate
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
	PaginationResponse
	Data []ProjectMember `json:"data"`
} // @name ProjectMemberListResponse

type BasicCountryInfo struct {
	ID   UUID   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
} // @name BasicCountryInfo

type UpdateProjectGeneralInfo struct {
	Name          string                `json:"name"`
	StartDate     *time.Time            `json:"startDate"`
	Country       *BasicCountryInfo     `json:"country"`
	Stacks        []Stack               `json:"stacks"`
	Function      ProjectFunction       `json:"function"`
	AuditNotionID string                `json:"auditNotionID"`
	BankAccount   *BasicBankAccountInfo `json:"bankAccount"`
	Client        *Client               `json:"client"`
	Organization  *Organization         `json:"organization"`
} // @name UpdateProjectGeneralInfo

type ProjectFunction string // @name ProjectFunction

const (
	ProjectFunctionDevelopment ProjectFunction = "development"
	ProjectFunctionLearning    ProjectFunction = "learning"
	ProjectFunctionTraining    ProjectFunction = "training"
	ProjectFunctionManagement  ProjectFunction = "management"
)

func (e ProjectFunction) IsValid() bool {
	switch e {
	case
		ProjectFunctionDevelopment,
		ProjectFunctionLearning,
		ProjectFunctionTraining,
		ProjectFunctionManagement:
		return true
	}
	return false
}

func (e ProjectFunction) String() string {
	return string(e)
}

type UpdateProjectGeneralInfoResponse struct {
	Data UpdateProjectGeneralInfo `json:"data"`
} // @name UpdateProjectGeneralInfoResponse

func ToUpdateProjectGeneralInfo(project *model.Project) UpdateProjectGeneralInfo {
	stacks := make([]Stack, 0, len(project.ProjectStacks))
	for _, v := range project.ProjectStacks {
		s := Stack{
			ID:     v.Stack.ID.String(),
			Name:   v.Stack.Name,
			Code:   v.Stack.Code,
			Avatar: v.Stack.Avatar,
		}
		stacks = append(stacks, s)
	}

	rs := UpdateProjectGeneralInfo{
		Name:      project.Name,
		StartDate: project.StartDate,
		Stacks:    stacks,
		Function:  ProjectFunction(project.Function),
	}

	if project.ProjectNotion != nil && !project.ProjectNotion.AuditNotionID.IsZero() {
		rs.AuditNotionID = project.ProjectNotion.AuditNotionID.String()
	}

	if project.Country != nil {
		rs.Country = &BasicCountryInfo{
			ID:   UUID(project.Country.ID),
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
		rs.Client = ToClient(project.Client)
	}

	return rs
}

type WorkUnitType string // @name WorkUnitType

const (
	WorkUnitTypeDevelopment WorkUnitType = "development"
	WorkUnitTypeManagement  WorkUnitType = "management"
	WorkUnitTypeTraining    WorkUnitType = "training"
	WorkUnitTypeLearning    WorkUnitType = "learning"
)

func (e WorkUnitType) IsValid() bool {
	switch e {
	case
		WorkUnitTypeDevelopment,
		WorkUnitTypeManagement,
		WorkUnitTypeTraining,
		WorkUnitTypeLearning:
		return true
	}
	return false
}

func (e WorkUnitType) String() string {
	return string(e)
}

type BasicProjectHeadInfo struct {
	EmployeeID     string          `json:"employeeID"`
	FullName       string          `json:"fullName"`
	DisplayName    string          `json:"displayName"`
	Avatar         string          `json:"avatar"`
	Position       HeadPosition    `json:"position"`
	Username       string          `json:"username"`
	CommissionRate decimal.Decimal `json:"commissionRate"`
} // @name BasicProjectHeadInfo

type HeadPosition string // @name HeadPosition

type UpdateProjectContactInfo struct {
	ClientEmail  []string               `json:"clientEmail"`
	ProjectEmail string                 `json:"projectEmail"`
	ProjectHead  []BasicProjectHeadInfo `json:"projectHead"`
} // @name UpdateProjectContactInfo

type UpdateProjectContactInfoResponse struct {
	Data UpdateProjectContactInfo `json:"data"`
} // @name UpdateProjectContactInfoResponse

func ToUpdateProjectContactInfo(project *model.Project, userInfo *model.CurrentLoggedUserInfo) UpdateProjectContactInfo {
	projectHeads := make([]BasicProjectHeadInfo, 0, len(project.Heads))
	for _, v := range project.Heads {
		ph := BasicProjectHeadInfo{
			EmployeeID:  v.Employee.ID.String(),
			FullName:    v.Employee.FullName,
			Avatar:      v.Employee.Avatar,
			DisplayName: v.Employee.DisplayName,
			Position:    HeadPosition(v.Position),
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
} // @name BasicProjectInfo

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
} // @name ProjectContentData

type ProjectContentDataResponse struct {
	Data *ProjectContentData `json:"data"`
} // @name ProjectContentDataResponse

func ToProjectContentData(url string) *ProjectContentData {
	return &ProjectContentData{
		Url: url,
	}
}
