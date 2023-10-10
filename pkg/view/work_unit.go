package view

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type BasicMember struct {
	EmployeeID  string `json:"employeeID"`
	FullName    string `json:"fullName"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Userame     string `json:"username"`
} // @name BasicMember

func toBasicMember(employee model.Employee) *BasicMember {
	return &BasicMember{
		EmployeeID:  employee.ID.String(),
		FullName:    employee.FullName,
		DisplayName: employee.DisplayName,
		Avatar:      employee.Avatar,
		Userame:     employee.Username,
	}
}

type WorkUnit struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	URL       string        `json:"url"`
	Members   []BasicMember `json:"members"`
	Stacks    []Stack       `json:"stacks"`
	Type      string        `json:"type"`
	Status    string        `json:"status"`
	ProjectID string        `json:"projectID"`
	Code      string        `json:"code"`
} // @name WorkUnit

func ToWorkUnit(workUnit *model.WorkUnit, projectCode string) WorkUnit {
	rs := WorkUnit{
		ID:        workUnit.ID.String(),
		Name:      workUnit.Name,
		Type:      workUnit.Type.String(),
		Status:    workUnit.Status.String(),
		URL:       workUnit.SourceURL,
		ProjectID: workUnit.ProjectID.String(),
		Code:      projectCode,
	}

	members := make([]BasicMember, 0, len(workUnit.WorkUnitMembers))
	for _, v := range workUnit.WorkUnitMembers {
		members = append(members, *toBasicMember(v.Employee))
	}
	rs.Members = members

	stacks := make([]Stack, 0, len(workUnit.WorkUnitStacks))
	for _, v := range workUnit.WorkUnitStacks {
		stack := Stack{
			ID:     v.Stack.ID.String(),
			Code:   v.Stack.Code,
			Name:   v.Stack.Name,
			Avatar: v.Stack.Avatar,
		}
		stacks = append(stacks, stack)
	}
	rs.Stacks = stacks

	return rs
}

type ListWorkUnitResponse struct {
	Data []WorkUnit `json:"data"`
} // @name ListWorkUnitResponse

type WorkUnitResponse struct {
	Data WorkUnit `json:"data"`
} // @name WorkUnitResponse

func ToWorkUnitList(workUnits []*model.WorkUnit, projectID string, projectCode string) []*WorkUnit {
	var rs []*WorkUnit

	for _, wu := range workUnits {
		newWorkUnit := &WorkUnit{
			ID:        wu.ID.String(),
			Name:      wu.Name,
			URL:       wu.SourceURL,
			Type:      wu.Type.String(),
			Status:    wu.Status.String(),
			ProjectID: projectID,
			Code:      projectCode,
		}

		for _, member := range wu.WorkUnitMembers {
			newWorkUnit.Members = append(newWorkUnit.Members, *toBasicMember(member.Employee))
		}

		for _, wStack := range wu.WorkUnitStacks {
			newWorkUnit.Stacks = append(newWorkUnit.Stacks, Stack{
				ID:     wStack.Stack.ID.String(),
				Name:   wStack.Stack.Name,
				Code:   wStack.Stack.Code,
				Avatar: wStack.Stack.Avatar,
			})
		}

		rs = append(rs, newWorkUnit)
	}

	return rs
}
