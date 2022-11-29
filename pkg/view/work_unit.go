package view

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type BasicMember struct {
	EmployeeID  string `json:"employeeID"`
	FullName    string `json:"fullName"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

type WorkUnit struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	URL       string        `json:"url"`
	Members   []BasicMember `json:"members"`
	Stacks    []MetaData    `json:"stacks"`
	Type      string        `json:"type"`
	Status    string        `json:"status"`
	ProjectID string        `json:"projectID"`
}

func ToWorkUnit(workUnit *model.WorkUnit) WorkUnit {
	rs := WorkUnit{
		ID:        workUnit.ID.String(),
		Name:      workUnit.Name,
		Type:      workUnit.Type.String(),
		Status:    workUnit.Status.String(),
		URL:       workUnit.SourceURL,
		ProjectID: workUnit.ProjectID.String(),
	}

	members := make([]BasicMember, 0, len(workUnit.WorkUnitMembers))
	for _, v := range workUnit.WorkUnitMembers {
		member := BasicMember{
			EmployeeID:  v.EmployeeID.String(),
			FullName:    v.Employee.FullName,
			DisplayName: v.Employee.DisplayName,
			Avatar:      v.Employee.Avatar,
		}
		members = append(members, member)
	}
	rs.Members = members

	stacks := make([]MetaData, 0, len(workUnit.WorkUnitStacks))
	for _, v := range workUnit.WorkUnitStacks {
		stack := MetaData{
			ID:   v.Stack.ID.String(),
			Code: v.Stack.Code,
			Name: v.Stack.Name,
		}
		stacks = append(stacks, stack)
	}
	rs.Stacks = stacks

	return rs
}

type ListWorkUnitResponse struct {
	Data []WorkUnit `json:"data"`
}

type WorkUnitResponse struct {
	Data WorkUnit `json:"data"`
}
