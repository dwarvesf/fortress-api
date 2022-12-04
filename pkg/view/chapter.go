package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type Chapter struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	LeadID string `json:"LeadID"`
}

func ToChapters(employeeChapters []model.EmployeeChapter) []Chapter {
	rs := make([]Chapter, 0, len(employeeChapters))
	for _, v := range employeeChapters {
		r := Chapter{
			ID:     v.Chapter.ID.String(),
			Code:   v.Chapter.Code,
			Name:   v.Chapter.Name,
			LeadID: v.Chapter.LeadID.String(),
		}
		rs = append(rs, r)
	}

	return rs
}
