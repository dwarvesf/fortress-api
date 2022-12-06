package metadata

import "github.com/dwarvesf/fortress-api/pkg/model"

// GetQuestionsInput input params for get questions api
type GetQuestionsInput struct {
	Category    model.EventType    `form:"category" json:"category" binding:"required"`
	Subcategory model.EventSubtype `form:"subcategory" json:"subcategory" binding:"required"`
}

// Validate check valid for values in input params
func (i GetQuestionsInput) Validate() error {
	if i.Category == "" || !i.Category.IsValid() {
		return ErrInvalidcategory
	}

	if i.Subcategory == "" || !i.Subcategory.IsValid() {
		return ErrInvalidSubcategory
	}

	return nil
}
