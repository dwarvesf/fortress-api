package request

import "errors"

type TopMemoAuthorsInput struct {
	Limit int `json:"limit"`
	Days  int `json:"days"`
}

func (i *TopMemoAuthorsInput) Validate() error {
	if i.Limit <= 0 {
		return errors.New("limit must be greater than 0")
	}
	if i.Days <= 0 {
		return errors.New("days must be greater than 0")
	}
	return nil
}
