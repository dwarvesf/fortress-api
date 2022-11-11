package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type ProjectData struct {
	model.BaseModel

	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Status    string          `json:"status"`
	StartDate *time.Time      `json:"startDate"`
	EndDate   *time.Time      `json:"endDate"`
	Members   []ProjectMember `json:"members"`
	Heads     []ProjectHead   `json:"heads"`
}

type ProjectMember struct {
	EmployeeID  string `json:"employeeID"`
	FullName    string `json:"fullName"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	IsLead      bool   `json:"isLead"`
}

type ProjectHead struct {
	EmployeeID  string `json:"employeeID"`
	FullName    string `json:"fullName"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Position    string `json:"position"`
}

func ToProjectData(projects []*model.Project) []ProjectData {
	var results = make([]ProjectData, 0, len(projects))

	for _, p := range projects {
		leads := map[string]string{}
		var heads = make([]ProjectHead, 0, len(p.Heads))

		for _, h := range p.Heads {
			if h.IsLead() {
				leads[h.EmployeeID.String()] = h.Position.String()
			}

			heads = append(heads, ProjectHead{
				EmployeeID:  h.EmployeeID.String(),
				FullName:    h.Employee.FullName,
				DisplayName: h.Employee.DisplayName,
				Avatar:      h.Employee.Avatar,
				Position:    h.Position.String(),
			})
		}

		var members = make([]ProjectMember, 0, len(p.Members))
		for _, m := range p.Members {
			_, isLead := leads[m.Employee.ID.String()]

			members = append(members, ProjectMember{
				EmployeeID:  m.ID.String(),
				FullName:    m.Employee.FullName,
				DisplayName: m.Employee.DisplayName,
				Avatar:      m.Employee.Avatar,
				IsLead:      isLead,
			})
		}

		d := ProjectData{
			BaseModel: p.BaseModel,
			Name:      p.Name,
			Type:      p.Type.String(),
			Status:    p.Status.String(),
			StartDate: p.StartDate,
			EndDate:   p.EndDate,
			Members:   members,
			Heads:     heads,
		}
		results = append(results, d)
	}

	return results
}

type ProjectListDataResponse struct {
	Data []ProjectData `json:"data"`
}
