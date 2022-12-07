package store

import (
	"github.com/dwarvesf/fortress-api/pkg/store/chapter"
	"github.com/dwarvesf/fortress-api/pkg/store/content"
	"github.com/dwarvesf/fortress-api/pkg/store/country"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/employeechapter"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeeventquestion"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeeventreviewer"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeeventtopic"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeposition"
	"github.com/dwarvesf/fortress-api/pkg/store/employeerole"
	"github.com/dwarvesf/fortress-api/pkg/store/employeestack"
	"github.com/dwarvesf/fortress-api/pkg/store/feedbackevent"
	"github.com/dwarvesf/fortress-api/pkg/store/permission"
	"github.com/dwarvesf/fortress-api/pkg/store/position"
	"github.com/dwarvesf/fortress-api/pkg/store/project"
	"github.com/dwarvesf/fortress-api/pkg/store/projecthead"
	"github.com/dwarvesf/fortress-api/pkg/store/projectmember"
	"github.com/dwarvesf/fortress-api/pkg/store/projectmemberposition"
	"github.com/dwarvesf/fortress-api/pkg/store/projectslot"
	"github.com/dwarvesf/fortress-api/pkg/store/projectslotposition"
	"github.com/dwarvesf/fortress-api/pkg/store/projectstack"
	"github.com/dwarvesf/fortress-api/pkg/store/question"
	"github.com/dwarvesf/fortress-api/pkg/store/role"
	"github.com/dwarvesf/fortress-api/pkg/store/seniority"
	"github.com/dwarvesf/fortress-api/pkg/store/stack"
	"github.com/dwarvesf/fortress-api/pkg/store/workunit"
	"github.com/dwarvesf/fortress-api/pkg/store/workunitmember"
	"github.com/dwarvesf/fortress-api/pkg/store/workunitstack"
)

type Store struct {
	Employee              employee.IStore
	Seniority             seniority.IStore
	Chapter               chapter.IStore
	Position              position.IStore
	Permission            permission.IStore
	Country               country.IStore
	Role                  role.IStore
	Project               project.IStore
	ProjectHead           projecthead.IStore
	ProjectMember         projectmember.IStore
	ProjectMemberPosition projectmemberposition.IStore
	ProjectSlot           projectslot.IStore
	ProjectSlotPosition   projectslotposition.IStore
	Stack                 stack.IStore
	EmployeePosition      employeeposition.IStore
	EmployeeRole          employeerole.IStore
	EmployeeStack         employeestack.IStore
	EmployeeChapter       employeechapter.IStore
	ProjectStack          projectstack.IStore
	Content               content.IStore
	WorkUnit              workunit.IStore
	WorkUnitMember        workunitmember.IStore
	WorkUnitStack         workunitstack.IStore
	EmployeeEventTopic    employeeeventtopic.IStore
	Question              question.IStore
	EmployeeEventQuestion employeeeventquestion.IStore
	EmployeeEventReviewer employeeeventreviewer.IStore
	FeedbackEvent         feedbackevent.IStore
}

func New() *Store {
	return &Store{
		Employee:              employee.New(),
		Seniority:             seniority.New(),
		Chapter:               chapter.New(),
		Position:              position.New(),
		Permission:            permission.New(),
		Country:               country.New(),
		Role:                  role.New(),
		Project:               project.New(),
		ProjectHead:           projecthead.New(),
		ProjectMember:         projectmember.New(),
		ProjectMemberPosition: projectmemberposition.New(),
		ProjectSlot:           projectslot.New(),
		ProjectSlotPosition:   projectslotposition.New(),
		Stack:                 stack.New(),
		EmployeePosition:      employeeposition.New(),
		EmployeeRole:          employeerole.New(),
		EmployeeStack:         employeestack.New(),
		EmployeeChapter:       employeechapter.New(),
		ProjectStack:          projectstack.New(),
		Content:               content.New(),
		WorkUnit:              workunit.New(),
		WorkUnitMember:        workunitmember.New(),
		WorkUnitStack:         workunitstack.New(),
		EmployeeEventTopic:    employeeeventtopic.New(),
		Question:              question.New(),
		EmployeeEventQuestion: employeeeventquestion.New(),
		EmployeeEventReviewer: employeeeventreviewer.New(),
		FeedbackEvent:         feedbackevent.New(),
	}
}
