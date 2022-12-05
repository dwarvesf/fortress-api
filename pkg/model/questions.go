package model

// Question model for questions table
type Question struct {
	BaseModel

	Type        string
	Category    string
	Subcategory string
	Code        string
	Content     string
}
