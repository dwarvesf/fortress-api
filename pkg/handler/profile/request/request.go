package request

import (
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/handler/profile/errs"
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
	Avatar                 string     `json:"avatar"`
	Address                string     `json:"address" binding:"required"`
	City                   string     `json:"city" binding:"required"`
	Country                string     `json:"country" binding:"required"`
	DateOfBirth            *time.Time `json:"dateOfBirth" binding:"required"`
	Gender                 string     `json:"gender" binding:"required"`
	Horoscope              string     `json:"horoscope" binding:"required"`
	MBTI                   string     `json:"mbti" binding:"required"`
	PhoneNumber            string     `json:"phoneNumber" binding:"required,max=18,min=8"`
	PlaceOfResidence       string     `json:"placeOfResidence" binding:"required"`
	PassportPhotoFront     string     `json:"passportPhotoFront"`
	PassportPhotoBack      string     `json:"passportPhotoBack"`
	IdentityCardPhotoFront string     `json:"identityCardPhotoFront"`
	IdentityCardPhotoBack  string     `json:"identityCardPhotoBack"`

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

func (i *SubmitOnboardingFormRequest) Validate() error {
	if i.DateOfBirth.After(time.Now()) {
		return errs.ErrInvalidDate
	}

	if i.PassportPhotoBack == "" || i.PassportPhotoFront == "" {
		if i.IdentityCardPhotoFront == "" || i.IdentityCardPhotoBack == "" {
			return errs.ErrMissingDocuments
		}
	}

	if i.IdentityCardPhotoFront == "" || i.IdentityCardPhotoBack == "" {
		if i.PassportPhotoBack == "" || i.PassportPhotoFront == "" {
			return errs.ErrMissingDocuments
		}
	}

	return nil
}

func (i *SubmitOnboardingFormRequest) ToEmployeeModel() *model.Employee {
	return &model.Employee{
		Avatar:                 i.Avatar,
		Address:                i.Address,
		City:                   i.City,
		Country:                i.Country,
		DateOfBirth:            i.DateOfBirth,
		Gender:                 i.Gender,
		Horoscope:              i.Horoscope,
		PassportPhotoFront:     strings.TrimSpace(i.PassportPhotoFront),
		PassportPhotoBack:      strings.TrimSpace(i.PassportPhotoBack),
		IdentityCardPhotoFront: strings.TrimSpace(i.IdentityCardPhotoFront),
		IdentityCardPhotoBack:  strings.TrimSpace(i.IdentityCardPhotoBack),
		LocalBranchName:        i.LocalBranchName,
		LocalBankBranch:        i.LocalBankBranch,
		LocalBankCurrency:      i.LocalBankCurrency,
		LocalBankNumber:        i.LocalBankNumber,
		LocalBankRecipientName: i.LocalBankRecipientName,
		MBTI:                   i.MBTI,
		PhoneNumber:            i.PhoneNumber,
		PlaceOfResidence:       i.PlaceOfResidence,
		WorkingStatus:          model.WorkingStatusProbation,
	}
}
