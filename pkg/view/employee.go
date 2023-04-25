package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// EmployeeData view for listing data
type EmployeeData struct {
	model.BaseModel

	// basic info
	FullName         string     `json:"fullName"`
	DisplayName      string     `json:"displayName"`
	TeamEmail        string     `json:"teamEmail"`
	PersonalEmail    string     `json:"personalEmail"`
	Avatar           string     `json:"avatar"`
	PhoneNumber      string     `json:"phoneNumber"`
	Address          string     `json:"address"`
	PlaceOfResidence string     `json:"placeOfResidence"`
	Country          string     `json:"country"`
	City             string     `json:"city"`
	MBTI             string     `json:"mbti"`
	Gender           string     `json:"gender"`
	Horoscope        string     `json:"horoscope"`
	DateOfBirth      *time.Time `json:"birthday"`
	Username         string     `json:"username"`
	GithubID         string     `json:"githubID"`
	NotionID         string     `json:"notionID"`
	NotionName       string     `json:"notionName"`
	DiscordID        string     `json:"discordID"`
	DiscordName      string     `json:"discordName"`
	LinkedInName     string     `json:"linkedInName"`

	// working info
	WorkingStatus model.WorkingStatus `json:"status"`
	JoinedDate    *time.Time          `json:"joinedDate"`
	LeftDate      *time.Time          `json:"leftDate"`

	Seniority          *model.Seniority      `json:"seniority"`
	LineManager        *BasicEmployeeInfo    `json:"lineManager"`
	ReferredBy         *BasicEmployeeInfo    `json:"referredBy"`
	Organizations      []Organization        `json:"organizations"`
	Positions          []Position            `json:"positions"`
	Stacks             []Stack               `json:"stacks"`
	Roles              []Role                `json:"roles"`
	Projects           []EmployeeProjectData `json:"projects"`
	Chapters           []Chapter             `json:"chapters"`
	Mentees            []*MenteeInfo         `json:"mentees"`
	BaseSalary         *BaseSalary           `json:"baseSalary"`
	WiseRecipientID    string                `json:"wiseRecipientID"`
	WiseAccountNumber  string                `json:"wiseAccountNumber"`
	WiseRecipientEmail string                `json:"wiseRecipientEmail"`
	WiseRecipientName  string                `json:"wiseRecipientName"`
	WiseCurrency       string                `json:"wiseCurrency"`
}

type MenteeInfo struct {
	ID          string           `json:"id"`
	FullName    string           `json:"fullName"`
	DisplayName string           `json:"displayName"`
	Avatar      string           `json:"avatar"`
	Username    string           `json:"username"`
	Seniority   *model.Seniority `json:"seniority"`
	Positions   []model.Position `json:"positions"`
}

type SocialAccount struct {
	GithubID     string `json:"githubID"`
	NotionID     string `json:"notionID"`
	NotionName   string `json:"notionName"`
	NotionEmail  string `json:"notionEmail"`
	DiscordID    string `json:"discordID"`
	DiscordName  string `json:"discordName"`
	LinkedInName string `json:"linkedInName"`
}

type BaseSalary struct {
	ID                    string            `json:"id"`
	EmployeeID            string            `json:"employee_id"`
	ContractAmount        int64             `json:"contract_amount"`
	CompanyAccountAmount  int64             `json:"company_account_amount"`
	PersonalAccountAmount int64             `json:"personal_account_amount"`
	InsuranceAmount       model.VietnamDong `json:"insurance_amount"`
	Type                  string            `json:"type"`
	Category              string            `json:"category"`
	CurrencyID            string            `json:"currency_id"`
	Currency              *Currency         `json:"currency"`
	Batch                 int               `json:"batch"`
	EffectiveDate         *time.Time        `json:"effective_date"`
}

func toMenteeInfo(employee model.Employee) *MenteeInfo {
	positions := make([]model.Position, 0, len(employee.EmployeePositions))
	for _, v := range employee.EmployeePositions {
		positions = append(positions, v.Position)
	}

	return &MenteeInfo{
		ID:          employee.ID.String(),
		FullName:    employee.FullName,
		DisplayName: employee.DisplayName,
		Avatar:      employee.Avatar,
		Username:    employee.Username,
		Seniority:   employee.Seniority,
		Positions:   positions,
	}
}

type EmployeeProjectData struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	DeploymentType string     `json:"deploymentType"`
	Status         string     `json:"status"`
	Positions      []Position `json:"positions"`
	Code           string     `json:"code"`
	Avatar         string     `json:"avatar"`
	StartDate      *time.Time `json:"startDate"`
	EndDate        *time.Time `json:"endDate"`
}

func ToEmployeeProjectDetailData(pm *model.ProjectMember, userInfo *model.CurrentLoggedUserInfo) EmployeeProjectData {
	rs := EmployeeProjectData{
		ID:        pm.ProjectID.String(),
		Name:      pm.Project.Name,
		Status:    model.ProjectMemberStatusActive.String(),
		Positions: ToProjectMemberPositions(pm.ProjectMemberPositions),
		Code:      pm.Project.Code,
		Avatar:    pm.Project.Avatar,
	}

	if !pm.IsActive() ||
		pm.Project.Status == model.ProjectStatusClosed ||
		pm.Project.Status == model.ProjectStatusPaused {
		rs.Status = model.ProjectMemberStatusInactive.String()
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadProjectsFullAccess) {
		rs.StartDate = pm.StartDate
		rs.EndDate = pm.EndDate
		rs.DeploymentType = pm.DeploymentType.String()
	}

	return rs
}

func ToEmployeeProjectData(pm *model.ProjectMember) EmployeeProjectData {
	return EmployeeProjectData{
		ID:             pm.ProjectID.String(),
		Name:           pm.Project.Name,
		DeploymentType: pm.DeploymentType.String(),
		Status:         pm.Project.Status.String(),
		Positions:      ToProjectMemberPositions(pm.ProjectMemberPositions),
		Code:           pm.Project.Code,
		Avatar:         pm.Project.Avatar,
		StartDate:      pm.StartDate,
		EndDate:        pm.EndDate,
	}
}

type UpdateGeneralInfoEmployeeData struct {
	model.BaseModel

	// basic info
	FullName      string             `json:"fullName"`
	TeamEmail     string             `json:"teamEmail"`
	PhoneNumber   string             `json:"phoneNumber"`
	GithubID      string             `json:"githubID"`
	NotionID      string             `json:"notionID"`
	NotionName    string             `json:"notionName"`
	NotionEmail   string             `json:"notionEmail"`
	LinkedinName  string             `json:"linkedInName"`
	DiscordID     string             `json:"discordID"`
	DiscordName   string             `json:"discordName"`
	DisplayName   string             `json:"displayName"`
	Organizations []Organization     `json:"organizations"`
	LineManager   *BasicEmployeeInfo `json:"lineManager"`
	ReferredBy    *BasicEmployeeInfo `json:"referredBy"`
}

type UpdateSkillEmployeeData struct {
	model.BaseModel

	Seniority *model.Seniority `json:"seniority"`
	Positions []model.Position `json:"positions"`
	Stacks    []model.Stack    `json:"stacks"`
	Chapters  []model.Chapter  `json:"chapters"`
}

type UpdatePersonalEmployeeData struct {
	model.BaseModel

	PersonalEmail    string     `json:"personalEmail"`
	Address          string     `json:"address"`
	PlaceOfResidence string     `json:"placeOfResidence"`
	Gender           string     `json:"gender"`
	DateOfBirth      *time.Time `json:"birthday"`
	Country          string     `json:"country"`
	City             string     `json:"city"`
}

type BasicEmployeeInfo struct {
	ID          string `json:"id"`
	FullName    string `json:"fullName"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Username    string `json:"username"`
}

type UpdateEmployeeStatusResponse struct {
	Data EmployeeData `json:"data"`
}

type EmployeeListDataResponse struct {
	Data []EmployeeData `json:"data"`
}

type UpdataEmployeeStatusResponse struct {
	Data EmployeeData `json:"data"`
}

type UpdateSkillsEmployeeResponse struct {
	Data UpdateSkillEmployeeData `json:"data"`
}

type UpdatePersonalEmployeeResponse struct {
	Data UpdatePersonalEmployeeData `json:"data"`
}

type UpdateGeneralEmployeeResponse struct {
	Data UpdateGeneralInfoEmployeeData `json:"data"`
}
type UpdateBaseSalaryResponse struct {
	Data BaseSalary `json:"data"`
}

func ToUpdatePersonalEmployeeData(employee *model.Employee) *UpdatePersonalEmployeeData {
	return &UpdatePersonalEmployeeData{
		BaseModel: model.BaseModel{
			ID:        employee.ID,
			CreatedAt: employee.CreatedAt,
			UpdatedAt: employee.UpdatedAt,
		},
		DateOfBirth:      employee.DateOfBirth,
		Gender:           employee.Gender,
		Address:          employee.Address,
		PlaceOfResidence: employee.PlaceOfResidence,
		PersonalEmail:    employee.PersonalEmail,
		Country:          employee.Country,
		City:             employee.City,
	}
}

func ToUpdateSkillEmployeeData(employee *model.Employee) *UpdateSkillEmployeeData {
	positions := make([]model.Position, 0, len(employee.EmployeePositions))
	for _, v := range employee.EmployeePositions {
		positions = append(positions, v.Position)
	}

	stacks := make([]model.Stack, 0, len(employee.EmployeeStacks))
	for _, v := range employee.EmployeeStacks {
		stacks = append(stacks, v.Stack)
	}

	chapters := make([]model.Chapter, 0, len(employee.EmployeeChapters))
	for _, v := range employee.EmployeeChapters {
		chapters = append(chapters, v.Chapter)
	}

	return &UpdateSkillEmployeeData{
		BaseModel: model.BaseModel{
			ID:        employee.ID,
			CreatedAt: employee.CreatedAt,
			UpdatedAt: employee.UpdatedAt,
		},
		Seniority: employee.Seniority,
		Positions: positions,
		Stacks:    stacks,
		Chapters:  chapters,
	}
}

func ToUpdateGeneralInfoEmployeeData(employee *model.Employee) *UpdateGeneralInfoEmployeeData {
	rs := &UpdateGeneralInfoEmployeeData{
		BaseModel: model.BaseModel{
			ID:        employee.ID,
			CreatedAt: employee.CreatedAt,
			UpdatedAt: employee.UpdatedAt,
		},
		FullName:      employee.FullName,
		TeamEmail:     employee.TeamEmail,
		PhoneNumber:   employee.PhoneNumber,
		GithubID:      employee.GithubID,
		NotionID:      employee.NotionID,
		NotionName:    employee.NotionName,
		NotionEmail:   employee.NotionEmail,
		DiscordID:     employee.DiscordID,
		DiscordName:   employee.DiscordName,
		LinkedinName:  employee.LinkedInName,
		DisplayName:   employee.DisplayName,
		Organizations: ToOrganizations(employee.EmployeeOrganizations),
	}

	if employee.LineManager != nil {
		rs.LineManager = toBasicEmployeeInfo(*employee.LineManager)
	}

	if !employee.ReferredBy.IsZero() {
		rs.ReferredBy = toBasicEmployeeInfo(*employee.Referrer)
	}

	return rs
}

// ToOneEmployeeData parse employee date to response data
func ToOneEmployeeData(employee *model.Employee, userInfo *model.CurrentLoggedUserInfo) *EmployeeData {
	employeeProjects := make([]EmployeeProjectData, 0, len(employee.ProjectMembers))
	for _, pm := range employee.ProjectMembers {
		// If logged user is working on the same project or user have permission to read active, show the project
		_, ok := userInfo.Projects[pm.ProjectID]
		if (ok || authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadProjectsReadActive)) &&
			pm.IsActive() && pm.Project.Status == model.ProjectStatusActive {
			employeeProjects = append(employeeProjects, ToEmployeeProjectDetailData(&pm, userInfo))
			continue
		}

		// If logged user have permission to read all projects, show the project
		if authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadProjectsFullAccess) {
			employeeProjects = append(employeeProjects, ToEmployeeProjectDetailData(&pm, userInfo))
		}
	}

	var lineManager, referrer *BasicEmployeeInfo
	if employee.LineManager != nil {
		lineManager = toBasicEmployeeInfo(*employee.LineManager)
	}
	if employee.Referrer != nil {
		referrer = toBasicEmployeeInfo(*employee.Referrer)
	}

	empSocialData := SocialAccount{}
	for _, sa := range employee.SocialAccounts {
		switch sa.Type {
		case model.SocialAccountTypeDiscord:
			empSocialData.DiscordID = sa.AccountID
			empSocialData.DiscordName = sa.Name
		case model.SocialAccountTypeGitHub:
			empSocialData.GithubID = sa.AccountID
		case model.SocialAccountTypeNotion:
			empSocialData.NotionID = sa.AccountID
			empSocialData.NotionName = sa.Name
		case model.SocialAccountTypeLinkedIn:
			empSocialData.LinkedInName = sa.AccountID
		}
	}

	rs := &EmployeeData{
		BaseModel: model.BaseModel{
			ID:        employee.ID,
			CreatedAt: employee.CreatedAt,
			UpdatedAt: employee.UpdatedAt,
		},

		FullName:    employee.FullName,
		DisplayName: employee.DisplayName,
		TeamEmail:   employee.TeamEmail,
		Avatar:      employee.Avatar,

		Gender:      employee.Gender,
		Horoscope:   employee.Horoscope,
		DateOfBirth: employee.DateOfBirth,

		Username:      employee.Username,
		WorkingStatus: employee.WorkingStatus,
		Seniority:     employee.Seniority,
		Projects:      employeeProjects,
		LineManager:   lineManager,
		Organizations: ToOrganizations(employee.EmployeeOrganizations),

		DiscordName: empSocialData.DiscordName,
		GithubID:    empSocialData.GithubID,

		Roles:     ToRoles(employee.EmployeeRoles),
		Positions: ToEmployeePositions(employee.EmployeePositions),
		Stacks:    ToEmployeeStacks(employee.EmployeeStacks),
		Chapters:  ToChapters(employee.EmployeeChapters),
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesBaseSalaryRead) {
		if !employee.BaseSalary.ID.IsZero() {
			rs.BaseSalary = ToBaseSalary(&employee.BaseSalary)
		}
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadGeneralInfoFullAccess) {
		rs.NotionID = empSocialData.NotionID
		rs.NotionName = empSocialData.NotionName
		rs.LinkedInName = empSocialData.LinkedInName
		rs.DiscordID = empSocialData.DiscordID
		rs.PhoneNumber = employee.PhoneNumber
		rs.JoinedDate = employee.JoinedDate
		rs.LeftDate = employee.LeftDate
		rs.ReferredBy = referrer
		rs.WiseRecipientID = employee.WiseRecipientID
		rs.WiseAccountNumber = employee.WiseAccountNumber
		rs.WiseRecipientEmail = employee.WiseRecipientEmail
		rs.WiseRecipientName = employee.WiseRecipientName
		rs.WiseCurrency = employee.WiseCurrency
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadPersonalInfoFullAccess) {
		rs.MBTI = employee.MBTI
		rs.PersonalEmail = employee.PersonalEmail
		rs.Address = employee.Address
		rs.PlaceOfResidence = employee.PlaceOfResidence
		rs.City = employee.City
		rs.Country = employee.Country
	}

	if len(employee.Mentees) > 0 {
		mentees := make([]*MenteeInfo, 0)
		for _, v := range employee.Mentees {
			mentees = append(mentees, toMenteeInfo(*v))
		}

		rs.Mentees = mentees
	}

	if employee.Seniority != nil {
		rs.Seniority = employee.Seniority
	}

	return rs
}

func ToEmployeeData(employee *model.Employee) *EmployeeData {
	employeeProjects := make([]EmployeeProjectData, 0)
	for _, v := range employee.ProjectMembers {
		employeeProjects = append(employeeProjects, ToEmployeeProjectData(&v))
	}

	var lineManager, referrer *BasicEmployeeInfo
	if employee.LineManager != nil {
		lineManager = toBasicEmployeeInfo(*employee.LineManager)
	}
	if employee.Referrer != nil {
		referrer = toBasicEmployeeInfo(*employee.Referrer)
	}

	rs := &EmployeeData{
		BaseModel: model.BaseModel{
			ID:        employee.ID,
			CreatedAt: employee.CreatedAt,
			UpdatedAt: employee.UpdatedAt,
		},
		FullName:         employee.FullName,
		DisplayName:      employee.DisplayName,
		TeamEmail:        employee.TeamEmail,
		PersonalEmail:    employee.PersonalEmail,
		Avatar:           employee.Avatar,
		PhoneNumber:      employee.PhoneNumber,
		Address:          employee.Address,
		PlaceOfResidence: employee.PlaceOfResidence,
		MBTI:             employee.MBTI,
		Gender:           employee.Gender,
		Horoscope:        employee.Horoscope,
		DateOfBirth:      employee.DateOfBirth,
		GithubID:         employee.GithubID,
		NotionID:         employee.NotionID,
		NotionName:       employee.NotionName,
		DiscordID:        employee.DiscordID,
		DiscordName:      employee.DiscordName,
		Username:         employee.Username,
		LinkedInName:     employee.LinkedInName,
		WorkingStatus:    employee.WorkingStatus,
		Seniority:        employee.Seniority,
		JoinedDate:       employee.JoinedDate,
		LeftDate:         employee.LeftDate,
		Organizations:    ToOrganizations(employee.EmployeeOrganizations),
		Projects:         employeeProjects,
		LineManager:      lineManager,
		ReferredBy:       referrer,
		Country:          employee.Country,
		City:             employee.City,
		Roles:            ToRoles(employee.EmployeeRoles),
		Positions:        ToEmployeePositions(employee.EmployeePositions),
		Stacks:           ToEmployeeStacks(employee.EmployeeStacks),
		Chapters:         ToChapters(employee.EmployeeChapters),
	}

	if len(employee.Mentees) > 0 {
		mentees := make([]*MenteeInfo, 0)
		for _, v := range employee.Mentees {
			mentees = append(mentees, toMenteeInfo(*v))
		}

		rs.Mentees = mentees
	}

	if employee.Seniority != nil {
		rs.Seniority = employee.Seniority
	}

	return rs
}

func ToEmployeeListData(employees []*model.Employee, userInfo *model.CurrentLoggedUserInfo) []EmployeeData {
	rs := make([]EmployeeData, 0, len(employees))
	for _, emp := range employees {
		empRes := ToOneEmployeeData(emp, userInfo)
		rs = append(rs, *empRes)
	}
	return rs
}

type EmployeeContentData struct {
	Url string `json:"url"`
}

type EmployeeContentDataResponse struct {
	Data *EmployeeContentData `json:"data"`
}

func ToEmployeeContentData(url string) *EmployeeContentData {
	return &EmployeeContentData{
		Url: url,
	}
}

func toBasicEmployeeInfo(employee model.Employee) *BasicEmployeeInfo {
	return &BasicEmployeeInfo{
		ID:          employee.ID.String(),
		FullName:    employee.FullName,
		DisplayName: employee.DisplayName,
		Avatar:      employee.Avatar,
		Username:    employee.Username,
	}
}

type LineManagersResponse struct {
	Data []BasicEmployeeInfo `json:"data"`
}

func ToBasicEmployees(employees []*model.Employee) []BasicEmployeeInfo {
	results := make([]BasicEmployeeInfo, 0, len(employees))

	for _, e := range employees {
		emp := toBasicEmployeeInfo(*e)
		results = append(results, *emp)
	}

	return results
}

func ToBaseSalary(bs *model.BaseSalary) *BaseSalary {
	if bs == nil {
		return nil
	}

	var currency *Currency
	if bs.Currency != nil {
		currency = &Currency{
			ID:     bs.Currency.ID.String(),
			Name:   bs.Currency.Name,
			Symbol: bs.Currency.Symbol,
			Locale: bs.Currency.Locale,
			Type:   bs.Currency.Type,
		}
	}
	return &BaseSalary{
		ID:                    bs.ID.String(),
		EmployeeID:            bs.EmployeeID.String(),
		ContractAmount:        bs.ContractAmount,
		CompanyAccountAmount:  bs.CompanyAccountAmount,
		PersonalAccountAmount: bs.PersonalAccountAmount,
		InsuranceAmount:       bs.InsuranceAmount,
		Type:                  bs.Type,
		Category:              bs.Category,
		CurrencyID:            bs.CurrencyID.String(),
		Currency:              currency,
		Batch:                 bs.Batch,
		EffectiveDate:         bs.EffectiveDate,
	}
}
