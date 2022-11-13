package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type ProjectData struct {
	model.BaseModel

	Name            string          `json:"name"`
	Type            string          `json:"type"`
	Status          string          `json:"status"`
	StartDate       *time.Time      `json:"startDate"`
	EndDate         *time.Time      `json:"endDate"`
	Members         []ProjectMember `json:"members"`
	TechnicalLead   []ProjectHead   `json:"technicalLeads"`
	AccountManager  *ProjectHead    `json:"accountManager"`
	SalePerson      *ProjectHead    `json:"salePerson"`
	DeliveryManager *ProjectHead    `json:"deliveryManager"`
}

type UpdatedProject struct {
	model.BaseModel

	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

type ProjectMember struct {
	EmployeeID  string `json:"employeeID"`
	FullName    string `json:"fullName"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Position    string `json:"position"`
	Status      string `json:"status"`
	IsLead      bool   `json:"isLead"`
}

type ProjectHead struct {
	EmployeeID  string `json:"employeeID"`
	FullName    string `json:"fullName"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

type UpdateProjectStatusResponse struct {
	Data UpdatedProject `json:"data"`
}

func ToUpdateProjectStatusResponse(p *model.Project) UpdatedProject {
	return UpdatedProject{
		BaseModel: p.BaseModel,
		Name:      p.Name,
		Type:      p.Type.String(),
		Status:    p.Status.String(),
		StartDate: p.StartDate,
		EndDate:   p.EndDate,
	}
}

func ToProjectData(projects []*model.Project) []ProjectData {
	var results = make([]ProjectData, 0, len(projects))

	for _, p := range projects {
		leads := map[string]string{}
		var technicalLeads = make([]ProjectHead, 0, len(p.Heads))
		var accountManager, salePerson, deliveryManager *ProjectHead
		for _, h := range p.Heads {
			head := ProjectHead{
				EmployeeID:  h.EmployeeID.String(),
				FullName:    h.Employee.FullName,
				DisplayName: h.Employee.DisplayName,
				Avatar:      h.Employee.Avatar,
			}

			if h.IsLead() {
				leads[h.EmployeeID.String()] = h.Position.String()
				technicalLeads = append(technicalLeads, head)
				continue
			}

			if h.IsAccountManager() {
				accountManager = &head
				continue
			}

			if h.IsSalePerson() {
				salePerson = &head
				continue
			}

			if h.IsDeliveryManager() {
				deliveryManager = &head
			}
		}

		var members = make([]ProjectMember, 0, len(p.Members))
		for _, m := range p.Members {
			_, isLead := leads[m.Employee.ID.String()]

			members = append(members, ProjectMember{
				EmployeeID:  m.ID.String(),
				FullName:    m.Employee.FullName,
				DisplayName: m.Employee.DisplayName,
				Avatar:      m.Employee.Avatar,
				Position:    m.Position,
				Status:      m.Status.String(),
				IsLead:      isLead,
			})
		}

		d := ProjectData{
			BaseModel:       p.BaseModel,
			Name:            p.Name,
			Type:            p.Type.String(),
			Status:          p.Status.String(),
			StartDate:       p.StartDate,
			EndDate:         p.EndDate,
			Members:         members,
			TechnicalLead:   technicalLeads,
			DeliveryManager: deliveryManager,
			SalePerson:      salePerson,
			AccountManager:  accountManager,
		}

		results = append(results, d)
	}

	return results
}

type ProjectListDataResponse struct {
	Data []ProjectData `json:"data"`
}

type EmployeeProjectData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func ToEmployeeProjectData(project *model.Project) EmployeeProjectData {
	return EmployeeProjectData{
		ID:   project.ID.String(),
		Name: project.Name,
	}
}
