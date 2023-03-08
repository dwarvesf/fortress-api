package model

type HiringType string

const (
	DirectHiringType     HiringType = "direct"
	ReferralHiringType   HiringType = "referral"
	InternshipHiringType HiringType = "internship"
)

type HiringRelationship struct {
	BaseModel

	UserID     UUID       `json:"user_id"`
	SupplierID UUID       `json:"supplier_id"`
	HiringType HiringType `json:"hiring_type"`

	Employee Employee `json:"user"`
}
