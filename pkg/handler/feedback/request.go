package feedback

import "github.com/dwarvesf/fortress-api/pkg/model"

type GetListFeedbackInput struct {
	model.Pagination

	Status string `json:"status" form:"status"`
}

func (i *GetListFeedbackInput) Validate() error {
	if i.Status != "" && !model.EventReviewerStatus(i.Status).IsValid() {
		return ErrInvalidReviewerStatus
	}

	return nil
}
