package view

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
)

// EmployeeData view for listing data
type EmployeeData struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`

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
	WorkingStatus WorkingStatus `json:"status"`
	JoinedDate    *time.Time    `json:"joinedDate"`
	LeftDate      *time.Time    `json:"leftDate"`

	Seniority          *Seniority            `json:"seniority"`
	LineManager        *BasicEmployeeInfo    `json:"lineManager"`
	ReferredBy         *BasicEmployeeInfo    `json:"referredBy"`
	Organizations      []Organization        `json:"organizations"`
	Positions          []Position            `json:"positions"`
	Projects           []EmployeeProjectData `json:"projects"`
	Stacks             []Stack               `json:"stacks"`
	Roles              []Role                `json:"roles"`
	Chapters           []Chapter             `json:"chapters"`
	Mentees            []*MenteeInfo         `json:"mentees"`
	BaseSalary         *BaseSalary           `json:"baseSalary"`
	WiseRecipientID    string                `json:"wiseRecipientID"`
	WiseAccountNumber  string                `json:"wiseAccountNumber"`
	WiseRecipientEmail string                `json:"wiseRecipientEmail"`
	WiseRecipientName  string                `json:"wiseRecipientName"`
	WiseCurrency       string                `json:"wiseCurrency"`
} // @name EmployeeData

type WorkingStatus string // @name WorkingStatus

type MMAScore struct {
	MasteryScore  decimal.Decimal `json:"masteryScore"`
	AutonomyScore decimal.Decimal `json:"autonomyScore"`
	MeaningScore  decimal.Decimal `json:"meaningScore"`
	RatedAt       *time.Time      `json:"ratedAt"`
}

type MenteeInfo struct {
	ID          string     `json:"id"`
	FullName    string     `json:"fullName"`
	DisplayName string     `json:"displayName"`
	Avatar      string     `json:"avatar"`
	Username    string     `json:"username"`
	Seniority   *Seniority `json:"seniority"`
	Positions   []Position `json:"positions"`
} // @name MenteeInfo

type SocialAccount struct {
	GithubID     string `json:"githubID"`
	NotionID     string `json:"notionID"`
	NotionName   string `json:"notionName"`
	NotionEmail  string `json:"notionEmail"`
	LinkedInName string `json:"linkedInName"`
}

type BaseSalary struct {
	ID                    string      `json:"id"`
	EmployeeID            string      `json:"employee_id"`
	ContractAmount        int64       `json:"contract_amount"`
	CompanyAccountAmount  int64       `json:"company_account_amount"`
	PersonalAccountAmount int64       `json:"personal_account_amount"`
	InsuranceAmount       VietnamDong `json:"insurance_amount"`
	Type                  string      `json:"type"`
	Category              string      `json:"category"`
	CurrencyID            string      `json:"currency_id"`
	Currency              *Currency   `json:"currency"`
	Batch                 int         `json:"batch"`
	EffectiveDate         *time.Time  `json:"effective_date"`
} // @name BaseSalary

type VietnamDong int64 // @name VietnamDong

func toMenteeInfo(employee model.Employee) *MenteeInfo {
	rs := &MenteeInfo{
		ID:          employee.ID.String(),
		FullName:    employee.FullName,
		DisplayName: employee.DisplayName,
		Avatar:      employee.Avatar,
		Username:    employee.Username,
		Positions:   ToEmployeePositions(employee.EmployeePositions),
	}

	if employee.Seniority != nil {
		s := ToSeniority(*employee.Seniority)
		rs.Seniority = &s
	}

	return rs
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
} // @name EmployeeProjectData

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
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`

	// basic info
	FullName      string             `json:"fullName"`
	TeamEmail     string             `json:"teamEmail"`
	PhoneNumber   string             `json:"phoneNumber"`
	GithubID      string             `json:"githubID"`
	NotionID      string             `json:"notionID"`
	NotionName    string             `json:"notionName"`
	NotionEmail   string             `json:"notionEmail"`
	LinkedInName  string             `json:"linkedInName"`
	DiscordID     string             `json:"discordID"`
	DiscordName   string             `json:"discordName"`
	DisplayName   string             `json:"displayName"`
	Organizations []Organization     `json:"organizations"`
	LineManager   *BasicEmployeeInfo `json:"lineManager"`
	ReferredBy    *BasicEmployeeInfo `json:"referredBy"`
} // @name UpdateGeneralInfoEmployeeData

type UpdateSkillEmployeeData struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`

	Seniority *Seniority `json:"seniority"`
	Positions []Position `json:"positions"`
	Stacks    []Stack    `json:"stacks"`
	Chapters  []Chapter  `json:"chapters"`
} // @name UpdateSkillEmployeeData

type UpdatePersonalEmployeeData struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`

	PersonalEmail    string     `json:"personalEmail"`
	Address          string     `json:"address"`
	PlaceOfResidence string     `json:"placeOfResidence"`
	Gender           string     `json:"gender"`
	DateOfBirth      *time.Time `json:"birthday"`
	Country          string     `json:"country"`
	City             string     `json:"city"`
} // @name UpdatePersonalEmployeeData

type BasicEmployeeInfo struct {
	ID          string `json:"id"`
	FullName    string `json:"fullName"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Username    string `json:"username"`
} // @name BasicEmployeeInfo

type UpdateEmployeeStatusResponse struct {
	Data EmployeeData `json:"data"`
} // @name UpdateEmployeeStatusResponse

type EmployeeListDataResponse struct {
	Data []EmployeeData `json:"data"`
} // @name EmployeeListDataResponse

type UpdataEmployeeStatusResponse struct {
	Data EmployeeData `json:"data"`
}

type UpdateSkillsEmployeeResponse struct {
	Data UpdateSkillEmployeeData `json:"data"`
} // @name UpdateSkillsEmployeeResponse

type UpdatePersonalEmployeeResponse struct {
	Data UpdatePersonalEmployeeData `json:"data"`
} // @name UpdatePersonalEmployeeResponse

type UpdateGeneralEmployeeResponse struct {
	Data UpdateGeneralInfoEmployeeData `json:"data"`
} // @name UpdateGeneralEmployeeResponse

type UpdateBaseSalaryResponse struct {
	Data BaseSalary `json:"data"`
} // @name UpdateBaseSalaryResponse

func ToUpdatePersonalEmployeeData(employee *model.Employee) *UpdatePersonalEmployeeData {
	return &UpdatePersonalEmployeeData{
		ID:               employee.ID.String(),
		CreatedAt:        employee.CreatedAt,
		UpdatedAt:        employee.UpdatedAt,
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

	rs := &UpdateSkillEmployeeData{
		ID:        employee.ID.String(),
		CreatedAt: employee.CreatedAt,
		UpdatedAt: employee.UpdatedAt,
		Positions: ToPositions(positions),
		Stacks:    ToEmployeeStacks(employee.EmployeeStacks),
		Chapters:  ToChapters(employee.EmployeeChapters),
	}

	if employee.Seniority != nil {
		s := ToSeniority(*employee.Seniority)
		rs.Seniority = &s
	}

	return rs
}

func ToUpdateGeneralInfoEmployeeData(employee *model.Employee) *UpdateGeneralInfoEmployeeData {
	rs := &UpdateGeneralInfoEmployeeData{
		ID:            employee.ID.String(),
		CreatedAt:     employee.CreatedAt,
		UpdatedAt:     employee.UpdatedAt,
		FullName:      employee.FullName,
		TeamEmail:     employee.TeamEmail,
		PhoneNumber:   employee.PhoneNumber,
		DisplayName:   employee.DisplayName,
		Organizations: ToOrganizations(employee.EmployeeOrganizations),
	}

	if len(employee.SocialAccounts) > 0 {
		for _, sa := range employee.SocialAccounts {
			switch sa.Type {
			case model.SocialAccountTypeGitHub:
				rs.GithubID = sa.AccountID
			case model.SocialAccountTypeNotion:
				rs.NotionID = sa.AccountID
				rs.NotionName = sa.Name
			case model.SocialAccountTypeLinkedIn:
				rs.LinkedInName = sa.Name
			}
		}
	}

	if employee.DiscordAccount != nil {
		rs.DiscordID = employee.DiscordAccount.DiscordID
		rs.DiscordName = employee.DiscordAccount.Username
	}

	if employee.LineManager != nil {
		rs.LineManager = toBasicEmployeeInfo(*employee.LineManager)
	}

	if !employee.ReferredBy.IsZero() {
		rs.ReferredBy = toBasicEmployeeInfo(*employee.Referrer)
	}

	return rs
}

type EmployeeDataResponse struct {
	Data *EmployeeData `json:"data"`
} // @name EmployeeDataResponse

// ToOneEmployeeData parse employee date to response data
func ToOneEmployeeData(employee *model.Employee, userInfo *model.CurrentLoggedUserInfo) *EmployeeData {
	employeeProjects := make([]EmployeeProjectData, 0, len(employee.ProjectMembers))
	for _, pm := range employee.ProjectMembers {
		if userInfo != nil {
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
		ID:          employee.ID.String(),
		CreatedAt:   employee.CreatedAt,
		UpdatedAt:   employee.UpdatedAt,
		FullName:    employee.FullName,
		DisplayName: employee.DisplayName,
		TeamEmail:   employee.TeamEmail,
		Avatar:      employee.Avatar,

		Gender:      employee.Gender,
		Horoscope:   employee.Horoscope,
		DateOfBirth: employee.DateOfBirth,

		Username:      employee.Username,
		WorkingStatus: WorkingStatus(employee.WorkingStatus.String()),
		Projects:      employeeProjects,
		LineManager:   lineManager,
		Organizations: ToOrganizations(employee.EmployeeOrganizations),

		GithubID: empSocialData.GithubID,

		Roles:     ToEmployeeRoles(employee.EmployeeRoles),
		Positions: ToEmployeePositions(employee.EmployeePositions),
		Stacks:    ToEmployeeStacks(employee.EmployeeStacks),
		Chapters:  ToChapters(employee.EmployeeChapters),
	}

	if employee.DiscordAccount != nil {
		rs.DiscordID = employee.DiscordAccount.DiscordID
		rs.DiscordName = employee.DiscordAccount.Username
	}

	if userInfo != nil && authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesBaseSalaryRead) {
		if !employee.BaseSalary.ID.IsZero() {
			rs.BaseSalary = ToBaseSalary(&employee.BaseSalary)
		}
	}

	if userInfo != nil && authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadGeneralInfoFullAccess) {
		rs.NotionID = empSocialData.NotionID
		rs.NotionName = empSocialData.NotionName
		rs.LinkedInName = empSocialData.LinkedInName
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

	if userInfo != nil && authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadPersonalInfoFullAccess) {
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
		s := ToSeniority(*employee.Seniority)
		rs.Seniority = &s
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
		ID:               employee.ID.String(),
		CreatedAt:        employee.CreatedAt,
		UpdatedAt:        employee.UpdatedAt,
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
		Username:         employee.Username,
		WorkingStatus:    WorkingStatus(employee.WorkingStatus.String()),
		JoinedDate:       employee.JoinedDate,
		LeftDate:         employee.LeftDate,
		Organizations:    ToOrganizations(employee.EmployeeOrganizations),
		Projects:         employeeProjects,
		LineManager:      lineManager,
		ReferredBy:       referrer,
		Country:          employee.Country,
		City:             employee.City,
		Roles:            ToEmployeeRoles(employee.EmployeeRoles),
		Positions:        ToEmployeePositions(employee.EmployeePositions),
		Stacks:           ToEmployeeStacks(employee.EmployeeStacks),
		Chapters:         ToChapters(employee.EmployeeChapters),
	}

	if len(employee.SocialAccounts) > 0 {
		for _, sa := range employee.SocialAccounts {
			switch sa.Type {
			case model.SocialAccountTypeGitHub:
				rs.GithubID = sa.AccountID
			case model.SocialAccountTypeNotion:
				rs.NotionID = sa.AccountID
				rs.NotionName = sa.Name
			case model.SocialAccountTypeLinkedIn:
				rs.LinkedInName = sa.Name
			}
		}
	}

	if employee.DiscordAccount != nil {
		rs.DiscordID = employee.DiscordAccount.DiscordID
		rs.DiscordName = employee.DiscordAccount.Username
	}

	if len(employee.Mentees) > 0 {
		mentees := make([]*MenteeInfo, 0)
		for _, v := range employee.Mentees {
			mentees = append(mentees, toMenteeInfo(*v))
		}

		rs.Mentees = mentees
	}

	if employee.Seniority != nil {
		s := ToSeniority(*employee.Seniority)
		rs.Seniority = &s
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
} // @name EmployeeContentData

type EmployeeContentDataResponse struct {
	Data *EmployeeContentData `json:"data"`
} // @name EmployeeContentDataResponse

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
} // @name LineManagersResponse

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
		InsuranceAmount:       VietnamDong(bs.InsuranceAmount),
		Type:                  bs.Type,
		Category:              bs.Category,
		CurrencyID:            bs.CurrencyID.String(),
		Currency:              currency,
		Batch:                 bs.Batch,
		EffectiveDate:         bs.EffectiveDate,
	}
}

type EmployeeInvitationData struct {
	ID                       string               `json:"id"`
	EmployeeID               string               `json:"employeeID"`
	InvitedBy                string               `json:"invitedBy"`
	IsCompleted              bool                 `json:"isCompleted"`
	IsInfoUpdated            bool                 `json:"isInfoUpdated"`
	IsDiscordRoleAssigned    bool                 `json:"isDiscordRoleAssigned"`
	IsBasecampAccountCreated bool                 `json:"isBasecampAccountCreated"`
	IsTeamEmailCreated       bool                 `json:"isTeamEmailCreated"`
	EmployeeData             *InvitedEmployeeInfo `json:"employee"`
} // @name EmployeeInvitationData

type InvitedEmployeeInfo struct {
	ID            string `json:"id"`
	FullName      string `json:"fullName"`
	DisplayName   string `json:"displayName"`
	Avatar        string `json:"avatar"`
	Username      string `json:"username"`
	TeamEmail     string `json:"teamEmail"`
	PersonalEmail string `json:"personalEmail"`
} // @name InvitedEmployeeInfo

type EmployeeInvitationResponse struct {
	Data *EmployeeInvitationData `json:"data"`
} // @name EmployeeInvitationResponse

func ToBasicEmployeeInvitationData(in *model.EmployeeInvitation) *EmployeeInvitationData {
	rs := &EmployeeInvitationData{
		ID:                       in.ID.String(),
		EmployeeID:               in.EmployeeID.String(),
		InvitedBy:                in.InvitedBy.String(),
		IsCompleted:              in.IsCompleted,
		IsInfoUpdated:            in.IsInfoUpdated,
		IsDiscordRoleAssigned:    in.IsDiscordRoleAssigned,
		IsBasecampAccountCreated: in.IsBasecampAccountCreated,
		IsTeamEmailCreated:       in.IsTeamEmailCreated,
	}

	if in.Employee != nil {
		rs.EmployeeData = &InvitedEmployeeInfo{
			ID:            in.Employee.ID.String(),
			FullName:      in.Employee.FullName,
			DisplayName:   in.Employee.DisplayName,
			Avatar:        in.Employee.Avatar,
			Username:      in.Employee.Username,
			TeamEmail:     in.Employee.TeamEmail,
			PersonalEmail: in.Employee.PersonalEmail,
		}
	}

	return rs
}

type EmployeeLocationListResponse struct {
	Data []EmployeeLocation `json:"data"`
} // @name EmployeeLocationListResponse

type EmployeeLocation struct {
	DiscordID   string          `json:"discordID"`
	FullName    string          `json:"fullName"`
	DisplayName string          `json:"displayName"`
	Avatar      string          `json:"avatar"`
	Chapters    []Chapter       `json:"chapters"`
	Address     EmployeeAddress `json:"address"`
} // @name EmployeeLocation

type EmployeeAddress struct {
	Address string `json:"address"`
	Country string `json:"country"`
	City    string `json:"city"`
	Lat     string `json:"lat"`
	Long    string `json:"long"`
} // @name EmployeeAddress

func ToEmployeesWithLocation(in []*model.Employee) []EmployeeLocation {
	rs := make([]EmployeeLocation, len(in))
	for i, v := range in {
		discordID := ""
		if v.DiscordAccount != nil {
			discordID = v.DiscordAccount.DiscordID
		}
		rs[i] = EmployeeLocation{
			DiscordID:   discordID,
			FullName:    v.FullName,
			DisplayName: v.DisplayName,
			Avatar:      v.Avatar,
			Chapters:    ToChapters(v.EmployeeChapters),
			Address: EmployeeAddress{
				Address: v.City + ", " + v.Country,
				Country: v.Country,
				City:    v.City,
				Lat:     v.Lat,
				Long:    v.Long,
			},
		}
	}
	return rs
}

// DiscordEmployeeData view for listing data
type DiscordEmployeeData struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`

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

	Seniority *model.Seniority      `json:"seniority"`
	Positions []Position            `json:"positions"`
	Stacks    []Stack               `json:"stacks"`
	Projects  []EmployeeProjectData `json:"projects"`

	MMAScore *MMAScore `json:"mmaScore"`
}

func ToDiscordEmployeeListData(employees []model.Employee, userInfo *model.CurrentLoggedUserInfo) []DiscordEmployeeData {
	rs := make([]DiscordEmployeeData, 0, len(employees))
	for _, emp := range employees {
		empRes := ToDiscordEmployeeDetail(&emp, userInfo)
		rs = append(rs, *empRes)
	}
	return rs
}
func ToDiscordEmployeeDetail(employee *model.Employee, userInfo *model.CurrentLoggedUserInfo) *DiscordEmployeeData {
	if employee == nil {
		return nil
	}

	employeeProjects := make([]EmployeeProjectData, 0, len(employee.ProjectMembers))
	for _, pm := range employee.ProjectMembers {
		if userInfo != nil {
			// If logged user is working on the same project or user have permission to read active, show the project
			if pm.IsActive() && pm.Project.Status == model.ProjectStatusActive {
				employeeProjects = append(employeeProjects, ToEmployeeProjectDetailData(&pm, userInfo))
				continue
			}
		}
	}

	empSocialData := SocialAccount{}
	for _, sa := range employee.SocialAccounts {
		switch sa.Type {
		case model.SocialAccountTypeGitHub:
			empSocialData.GithubID = sa.AccountID
		case model.SocialAccountTypeNotion:
			empSocialData.NotionID = sa.AccountID
			empSocialData.NotionName = sa.Name
		case model.SocialAccountTypeLinkedIn:
			empSocialData.LinkedInName = sa.AccountID
		}
	}

	rs := &DiscordEmployeeData{
		ID:        employee.ID.String(),
		CreatedAt: employee.CreatedAt,
		UpdatedAt: employee.UpdatedAt,

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

		GithubID: empSocialData.GithubID,

		Positions: ToEmployeePositions(employee.EmployeePositions),
		Stacks:    ToEmployeeStacks(employee.EmployeeStacks),
	}

	if employee.DiscordAccount != nil {
		rs.DiscordID = employee.DiscordAccount.DiscordID
		rs.DiscordName = employee.DiscordAccount.Username
	}

	rs.NotionID = empSocialData.NotionID
	rs.NotionName = empSocialData.NotionName
	rs.LinkedInName = empSocialData.LinkedInName
	rs.PhoneNumber = employee.PhoneNumber
	rs.JoinedDate = employee.JoinedDate

	rs.MBTI = employee.MBTI
	rs.PersonalEmail = employee.PersonalEmail
	rs.Address = employee.Address
	rs.PlaceOfResidence = employee.PlaceOfResidence
	rs.City = employee.City
	rs.Country = employee.Country

	if employee.Seniority != nil {
		rs.Seniority = employee.Seniority
	}

	if len(employee.EmployeeMMAScores) > 0 {
		rs.MMAScore = &MMAScore{
			MasteryScore:  employee.EmployeeMMAScores[0].MasteryScore,
			AutonomyScore: employee.EmployeeMMAScores[0].AutonomyScore,
			MeaningScore:  employee.EmployeeMMAScores[0].MeaningScore,
			RatedAt:       employee.EmployeeMMAScores[0].RatedAt,
		}
	}

	return rs
}

type EmployeeMMAScore struct {
	EmployeeID    string          `json:"employeeID"`
	FullName      string          `json:"fullName"`
	MMAID         string          `json:"mmaID"`
	MasteryScore  decimal.Decimal `json:"masteryScore"`
	AutonomyScore decimal.Decimal `json:"autonomyScore"`
	MeaningScore  decimal.Decimal `json:"meaningScore"`
	RatedAt       *time.Time      `json:"ratedAt"`
}

func ToEmployeesWithMMAScore(in []model.EmployeeMMAScoreData) []EmployeeMMAScore {
	rs := make([]EmployeeMMAScore, len(in))
	for i, v := range in {
		rs[i] = EmployeeMMAScore{
			EmployeeID:    v.EmployeeID.String(),
			FullName:      v.FullName,
			MMAID:         v.MMAID.String(),
			MasteryScore:  v.MasteryScore,
			AutonomyScore: v.AutonomyScore,
			MeaningScore:  v.MeaningScore,
			RatedAt:       v.RatedAt,
		}
	}

	return rs
}
