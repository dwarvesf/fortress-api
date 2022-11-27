package model

type Currency struct {
	BaseModel

	Name   string
	Symbol string
	Locale string
	Type   string
}
