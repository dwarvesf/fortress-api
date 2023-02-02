package store

import (
	"github.com/dwarvesf/fortress-api/pkg/store/actionitem"
	"github.com/dwarvesf/fortress-api/pkg/store/actionitemsnapshot"
	"github.com/dwarvesf/fortress-api/pkg/store/audit"
	"github.com/dwarvesf/fortress-api/pkg/store/auditactionitem"
	"github.com/dwarvesf/fortress-api/pkg/store/auditcycle"
	"github.com/dwarvesf/fortress-api/pkg/store/audititem"
	"github.com/dwarvesf/fortress-api/pkg/store/auditparticipant"
	"github.com/dwarvesf/fortress-api/pkg/store/chapter"
	"github.com/dwarvesf/fortress-api/pkg/store/content"
	"github.com/dwarvesf/fortress-api/pkg/store/country"
	"github.com/dwarvesf/fortress-api/pkg/store/dashboard"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/employeechapter"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeeventquestion"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeeventreviewer"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeeventtopic"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeorganization"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeposition"
	"github.com/dwarvesf/fortress-api/pkg/store/employeerole"
	"github.com/dwarvesf/fortress-api/pkg/store/employeestack"
	"github.com/dwarvesf/fortress-api/pkg/store/feedbackevent"
	"github.com/dwarvesf/fortress-api/pkg/store/organization"
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
	"github.com/dwarvesf/fortress-api/pkg/store/socialaccount"
	"github.com/dwarvesf/fortress-api/pkg/store/stack"
	"github.com/dwarvesf/fortress-api/pkg/store/valuation"
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
	FeedbackEvent         feedbackevent.IStore
	EmployeeEventReviewer employeeeventreviewer.IStore
	Dashboard             dashboard.IStore
	Valuation             valuation.IStore
	AuditCycle            auditcycle.IStore
	Audit                 audit.IStore
	ActionItem            actionitem.IStore
	AuditItem             audititem.IStore
	AuditParticipant      auditparticipant.IStore
	Organization          organization.IStore
	EmployeeOrganization  employeeorganization.IStore
	AuditActionItem       auditactionitem.IStore
	ActionItemSnapshot    actionitemsnapshot.IStore
	SocialAccount         socialaccount.IStore
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
		FeedbackEvent:         feedbackevent.New(),
		EmployeeEventReviewer: employeeeventreviewer.New(),
		Dashboard:             dashboard.New(),
		Valuation:             valuation.New(),
		AuditCycle:            auditcycle.New(),
		Audit:                 audit.New(),
		ActionItem:            actionitem.New(),
		AuditItem:             audititem.New(),
		AuditParticipant:      auditparticipant.New(),
		Organization:          organization.New(),
		EmployeeOrganization:  employeeorganization.New(),
		AuditActionItem:       auditactionitem.New(),
		ActionItemSnapshot:    actionitemsnapshot.New(),
		SocialAccount:         socialaccount.New(),
	}
}
