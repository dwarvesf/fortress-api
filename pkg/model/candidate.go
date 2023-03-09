package model

import (
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
)

type CandidateStatus string

const (
	ApproachCandidateStatus CandidateStatus = "approach"
	OfferedCandidateStatus  CandidateStatus = "offered"
	FailedCandidateStatus   CandidateStatus = "failed"
	HiredCandidateStatus    CandidateStatus = "hired"
	RejectCandidateStatus   CandidateStatus = "reject"
)

type Candidate struct {
	BaseModel

	Name              string          `json:"name"`
	Email             string          `json:"email"`
	Detail            string          `json:"detail"`
	Note              string          `json:"-" gorm:"-"`
	CVUrl             string          `json:"cv_url"`
	Role              string          `json:"role"`
	Source            string          `json:"source"`
	Type              HiringType      `json:"type"`
	Status            CandidateStatus `json:"status"`
	Phone             string          `json:"phone"`
	CCAT              int             `gorm:"column:CCAT" json:"CCAT"`
	EPP               int             `gorm:"column:EPP" json:"EPP"`
	IsReferral        bool            `json:"is_referral"`
	ReferralInfo      JSON            `json:"referral_info"`
	BasecampTodoID    int             `json:"basecamp_todo_id"`
	OfferSalary       int             `json:"offer_salary"`
	OfferStartDate    *time.Time      `json:"offer_start_date"`
	ProbationDuration int             `json:"probation_duration"`
	IsEmailSent       bool            `json:"is_email_sent"`
	OnboardTodoID     int             `json:"onboard_todo_id"`

	PdfFile          []byte `json:"-" gorm:"-"`
	GroupRole        string `json:"-" gorm:"-"`
	DisplayName      string `json:"-" gorm:"-"`
	DisplayStartDate string `json:"-" gorm:"-"`
	DisplaySalary    string `json:"-" gorm:"-"`
}

type ReferralInfo struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

func (r Candidate) FindHiringInCharge() int {
	switch r.Role {
	case "Backend":
		return consts.HuyNguyenBasecampID
	case "Frontend":
		return consts.HuyGiangBasecampID
	case "QA/QC":
		return consts.PhuongTruongBasecampID
	case "iOs", "MacOS":
		return consts.TrungPhanBasecampID
	case "Android":
		return consts.ThanhNguyenBasecampID
	case "Sales", "Client Partner":
		return consts.NamTranBasecampID
	case "Designer", "Ventures Designer", "Visual Designer":
		return consts.KhaiLeBasecampID
	}
	return consts.HuyNguyenBasecampID
}

func GroupRole(role string) string {
	switch role {
	case "Golang":
		return "Backend"
	case "React", "Vue":
		return "Frontend"
	case "undefined":
		return "Other"
	default:
		return role
	}
}
func DisplayRole(role string) string {
	switch role {
	case "Golang", "Backend":
		return "Back-end Engineer"
	case "React", "Vue", "Frontend":
		return "Front-end Engineer"
	default:
		return role
	}
}

func DisplayName(name string) string {
	parts := strings.Split(name, " ")
	return parts[len(parts)-1]
}
