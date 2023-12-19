package model

type Bank struct {
	BaseModel

	Name      string
	Code      string
	Bin       string
	ShortName string
	Logo      string
	SwiftCode string
}
