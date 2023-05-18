package request

import (
	"strings"
	"time"

	"github.com/mozillazg/go-unidecode"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// UpdateInfoInput input model for update profile
type UpdateInfoInput struct {
	PersonalEmail      string `form:"personalEmail" json:"personalEmail" binding:"required,email"`
	PhoneNumber        string `form:"phoneNumber" json:"phoneNumber" binding:"required,max=18,min=8"`
	PlaceOfResidence   string `form:"placeOfResidence" json:"placeOfResidence" binding:"required"`
	Address            string `form:"address" json:"address"`
	Country            string `form:"country" json:"country"`
	City               string `form:"city" json:"city"`
	GithubID           string `form:"githubID" json:"githubID"`
	NotionID           string `form:"notionID" json:"notionID"`
	NotionName         string `form:"notionName" json:"notionName"`
	NotionEmail        string `form:"notionEmail" json:"notionEmail"`
	DiscordName        string `form:"discordName" json:"discordName"`
	LinkedInName       string `form:"linkedInName" json:"linkedInName"`
	WiseRecipientID    string `form:"wiseRecipientID" json:"wiseRecipientID"`
	WiseRecipientEmail string `form:"wiseRecipientEmail" json:"wiseRecipientEmail" binding:"email"`
	WiseRecipientName  string `form:"wiseRecipientName" json:"wiseRecipientName"`
	WiseAccountNumber  string `form:"wiseAccountNumber" json:"wiseAccountNumber"`
	WiseCurrency       string `form:"wiseCurrency" json:"wiseCurrency"`
}

func (i UpdateInfoInput) ToEmployeeModel(employee *model.Employee) {
	employee.PersonalEmail = i.PersonalEmail
	employee.PhoneNumber = i.PhoneNumber
	employee.PlaceOfResidence = i.PlaceOfResidence
	employee.Address = i.Address
	employee.City = i.City
	employee.Country = i.Country

	if i.GithubID != "" {
		employee.GithubID = i.GithubID
	}

	if i.NotionID != "" {
		employee.NotionID = i.NotionID
	}

	if i.NotionName != "" {
		employee.NotionName = i.NotionName
	}

	if i.NotionEmail != "" {
		employee.NotionEmail = i.NotionEmail
	}

	if i.DiscordName != "" {
		employee.DiscordName = i.DiscordName
	}

	if i.LinkedInName != "" {
		employee.LinkedInName = i.LinkedInName
	}

	if strings.TrimSpace(i.WiseRecipientID) != "" {
		employee.WiseRecipientID = i.WiseRecipientID
	}
	if strings.TrimSpace(i.WiseRecipientEmail) != "" {
		employee.WiseRecipientEmail = i.WiseRecipientEmail
	}
	if strings.TrimSpace(i.WiseRecipientName) != "" {
		employee.WiseRecipientName = i.WiseRecipientName
	}
	if strings.TrimSpace(i.WiseAccountNumber) != "" {
		employee.WiseAccountNumber = i.WiseAccountNumber
	}
	if strings.TrimSpace(i.WiseCurrency) != "" {
		employee.WiseCurrency = i.WiseCurrency
	}
}

type SubmitOnboardingFormRequest struct {
	FullName  string `json:"-"`
	TeamEmail string `json:"-"`

	Address          string     `json:"address" binding:"required"`
	City             string     `json:"city" binding:"required"`
	Country          string     `json:"country" binding:"required"`
	DateOfBirth      *time.Time `json:"dateOfBirth" binding:"required"`
	Gender           string     `json:"gender" binding:"required"`
	Horoscope        string     `json:"horoscope" binding:"required"`
	MBTI             string     `json:"mbti" binding:"required"`
	PhoneNumber      string     `json:"phoneNumber" binding:"required,max=18,min=8"`
	PlaceOfResidence string     `json:"placeOfResidence" binding:"required"`

	LocalBankBranch        string `json:"localBankBranch" binding:"required"`
	LocalBankCurrency      string `json:"localBankCurrency" binding:"required"`
	LocalBankNumber        string `json:"localBankNumber" binding:"required"`
	LocalBankRecipientName string `json:"localBankRecipientName" binding:"required"`
	LocalBranchName        string `json:"localBranchName" binding:"required"`

	DiscordName  string `json:"discordName" binding:"required"`
	GithubID     string `json:"githubID"`
	LinkedInName string `json:"linkedInName"`
	NotionName   string `json:"notionName"`
}

func (i *SubmitOnboardingFormRequest) ToEmployeeModel(teamEmail string) *model.Employee {
	if teamEmail == "" {
		teamEmail = convertName(i.FullName) + "@d.foundation"
	}

	return &model.Employee{
		Address:                i.Address,
		City:                   i.City,
		Country:                i.Country,
		DateOfBirth:            i.DateOfBirth,
		Gender:                 i.Gender,
		Horoscope:              i.Horoscope,
		LocalBranchName:        i.LocalBranchName,
		LocalBankBranch:        i.LocalBankBranch,
		LocalBankCurrency:      i.LocalBankCurrency,
		LocalBankNumber:        i.LocalBankNumber,
		LocalBankRecipientName: i.LocalBankRecipientName,
		MBTI:                   i.MBTI,
		PhoneNumber:            i.PhoneNumber,
		PlaceOfResidence:       i.PlaceOfResidence,
		TeamEmail:              teamEmail,
		WorkingStatus:          model.WorkingStatusProbation,
	}
}

func convertName(fullName string) string {
	fullName = strings.TrimSpace(unidecode.Unidecode(fullName))
	nameParts := strings.Fields(fullName)

	numNameParts := len(nameParts)
	if numNameParts == 1 {
		return strings.ToLower(nameParts[0])
	}

	firstNameInitial := strings.ToLower(string(nameParts[0][0]))

	var middleNameInitial string
	if numNameParts > 2 {
		middleNameInitial = strings.ToLower(string(nameParts[1][0]))
	} else {
		middleNameInitial = ""
	}

	lastName := strings.ToLower(nameParts[numNameParts-1])

	return lastName + firstNameInitial + middleNameInitial
}
