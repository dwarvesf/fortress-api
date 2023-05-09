package store

import (
	"github.com/dwarvesf/fortress-api/pkg/store/accounting"
	"github.com/dwarvesf/fortress-api/pkg/store/actionitem"
	"github.com/dwarvesf/fortress-api/pkg/store/actionitemsnapshot"
	"github.com/dwarvesf/fortress-api/pkg/store/apikey"
	"github.com/dwarvesf/fortress-api/pkg/store/apikeyrole"
	"github.com/dwarvesf/fortress-api/pkg/store/audit"
	"github.com/dwarvesf/fortress-api/pkg/store/auditactionitem"
	"github.com/dwarvesf/fortress-api/pkg/store/auditcycle"
	"github.com/dwarvesf/fortress-api/pkg/store/audititem"
	"github.com/dwarvesf/fortress-api/pkg/store/auditparticipant"
	"github.com/dwarvesf/fortress-api/pkg/store/bankaccount"
	"github.com/dwarvesf/fortress-api/pkg/store/basesalary"
	"github.com/dwarvesf/fortress-api/pkg/store/bonus"
	"github.com/dwarvesf/fortress-api/pkg/store/cachedpayroll"
	"github.com/dwarvesf/fortress-api/pkg/store/chapter"
	"github.com/dwarvesf/fortress-api/pkg/store/client"
	"github.com/dwarvesf/fortress-api/pkg/store/clientcontact"
	"github.com/dwarvesf/fortress-api/pkg/store/commission"
	"github.com/dwarvesf/fortress-api/pkg/store/content"
	"github.com/dwarvesf/fortress-api/pkg/store/country"
	"github.com/dwarvesf/fortress-api/pkg/store/currency"
	"github.com/dwarvesf/fortress-api/pkg/store/dashboard"
	"github.com/dwarvesf/fortress-api/pkg/store/discordtemplate"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/employeechapter"
	"github.com/dwarvesf/fortress-api/pkg/store/employeecommission"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeeventquestion"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeeventreviewer"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeeventtopic"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeorganization"
	"github.com/dwarvesf/fortress-api/pkg/store/employeeposition"
	"github.com/dwarvesf/fortress-api/pkg/store/employeerole"
	"github.com/dwarvesf/fortress-api/pkg/store/employeestack"
	"github.com/dwarvesf/fortress-api/pkg/store/expense"
	"github.com/dwarvesf/fortress-api/pkg/store/feedbackevent"
	"github.com/dwarvesf/fortress-api/pkg/store/invoice"
	"github.com/dwarvesf/fortress-api/pkg/store/invoicenumbercaching"
	"github.com/dwarvesf/fortress-api/pkg/store/operationalservice"
	"github.com/dwarvesf/fortress-api/pkg/store/organization"
	"github.com/dwarvesf/fortress-api/pkg/store/payroll"
	"github.com/dwarvesf/fortress-api/pkg/store/permission"
	"github.com/dwarvesf/fortress-api/pkg/store/position"
	"github.com/dwarvesf/fortress-api/pkg/store/project"
	"github.com/dwarvesf/fortress-api/pkg/store/projectcommissionconfig"
	"github.com/dwarvesf/fortress-api/pkg/store/projecthead"
	"github.com/dwarvesf/fortress-api/pkg/store/projectmember"
	"github.com/dwarvesf/fortress-api/pkg/store/projectmemberposition"
	"github.com/dwarvesf/fortress-api/pkg/store/projectnotion"
	"github.com/dwarvesf/fortress-api/pkg/store/projectslot"
	"github.com/dwarvesf/fortress-api/pkg/store/projectslotposition"
	"github.com/dwarvesf/fortress-api/pkg/store/projectstack"
	"github.com/dwarvesf/fortress-api/pkg/store/question"
	"github.com/dwarvesf/fortress-api/pkg/store/recruitment"
	"github.com/dwarvesf/fortress-api/pkg/store/role"
	"github.com/dwarvesf/fortress-api/pkg/store/schedule"
	"github.com/dwarvesf/fortress-api/pkg/store/seniority"
	"github.com/dwarvesf/fortress-api/pkg/store/socialaccount"
	"github.com/dwarvesf/fortress-api/pkg/store/stack"
	"github.com/dwarvesf/fortress-api/pkg/store/valuation"
	"github.com/dwarvesf/fortress-api/pkg/store/workunit"
	"github.com/dwarvesf/fortress-api/pkg/store/workunitmember"
	"github.com/dwarvesf/fortress-api/pkg/store/workunitstack"
)

type Store struct {
	Accounting              accounting.IStore
	ActionItem              actionitem.IStore
	ActionItemSnapshot      actionitemsnapshot.IStore
	APIKey                  apikey.IStore
	APIKeyRole              apikeyrole.IStore
	Audit                   audit.IStore
	AuditActionItem         auditactionitem.IStore
	AuditCycle              auditcycle.IStore
	AuditItem               audititem.IStore
	AuditParticipant        auditparticipant.IStore
	BankAccount             bankaccount.IStore
	BaseSalary              basesalary.IStore
	Bonus                   bonus.IStore
	CachedPayroll           cachedpayroll.IStore
	Chapter                 chapter.IStore
	Client                  client.IStore
	ClientContact           clientcontact.IStore
	Commission              commission.IStore
	Content                 content.IStore
	Country                 country.IStore
	Currency                currency.IStore
	Dashboard               dashboard.IStore
	Employee                employee.IStore
	EmployeeChapter         employeechapter.IStore
	EmployeeCommission      employeecommission.IStore
	EmployeeEventQuestion   employeeeventquestion.IStore
	EmployeeEventReviewer   employeeeventreviewer.IStore
	EmployeeEventTopic      employeeeventtopic.IStore
	EmployeeOrganization    employeeorganization.IStore
	EmployeePosition        employeeposition.IStore
	EmployeeRole            employeerole.IStore
	EmployeeStack           employeestack.IStore
	Expense                 expense.IStore
	FeedbackEvent           feedbackevent.IStore
	Invoice                 invoice.IStore
	InvoiceNumberCaching    invoicenumbercaching.IStore
	Organization            organization.IStore
	Payroll                 payroll.IStore
	Permission              permission.IStore
	Position                position.IStore
	Project                 project.IStore
	ProjectCommissionConfig projectcommissionconfig.IStore
	ProjectHead             projecthead.IStore
	ProjectMember           projectmember.IStore
	ProjectMemberPosition   projectmemberposition.IStore
	ProjectNotion           projectnotion.IStore
	ProjectSlot             projectslot.IStore
	ProjectSlotPosition     projectslotposition.IStore
	ProjectStack            projectstack.IStore
	Question                question.IStore
	Recruitment             recruitment.IStore
	Role                    role.IStore
	Schedule                schedule.IStore
	Seniority               seniority.IStore
	SocialAccount           socialaccount.IStore
	Stack                   stack.IStore
	Valuation               valuation.IStore
	WorkUnit                workunit.IStore
	WorkUnitMember          workunitmember.IStore
	WorkUnitStack           workunitstack.IStore
	OperationalService      operationalservice.IStore
	DiscordLogTemplate      discordtemplate.IStore
}

func New() *Store {
	return &Store{
		Accounting:              accounting.New(),
		ActionItem:              actionitem.New(),
		ActionItemSnapshot:      actionitemsnapshot.New(),
		APIKey:                  apikey.New(),
		APIKeyRole:              apikeyrole.New(),
		Audit:                   audit.New(),
		AuditActionItem:         auditactionitem.New(),
		AuditCycle:              auditcycle.New(),
		AuditItem:               audititem.New(),
		AuditParticipant:        auditparticipant.New(),
		BankAccount:             bankaccount.New(),
		BaseSalary:              basesalary.New(),
		Bonus:                   bonus.New(),
		CachedPayroll:           cachedpayroll.New(),
		Chapter:                 chapter.New(),
		Client:                  client.New(),
		ClientContact:           clientcontact.New(),
		Commission:              commission.New(),
		Content:                 content.New(),
		Country:                 country.New(),
		Currency:                currency.New(),
		Dashboard:               dashboard.New(),
		Employee:                employee.New(),
		EmployeeChapter:         employeechapter.New(),
		EmployeeCommission:      employeecommission.New(),
		EmployeeEventQuestion:   employeeeventquestion.New(),
		EmployeeEventReviewer:   employeeeventreviewer.New(),
		EmployeeEventTopic:      employeeeventtopic.New(),
		EmployeeOrganization:    employeeorganization.New(),
		EmployeePosition:        employeeposition.New(),
		EmployeeRole:            employeerole.New(),
		EmployeeStack:           employeestack.New(),
		Expense:                 expense.New(),
		FeedbackEvent:           feedbackevent.New(),
		Invoice:                 invoice.New(),
		InvoiceNumberCaching:    invoicenumbercaching.New(),
		Organization:            organization.New(),
		Payroll:                 payroll.New(),
		Permission:              permission.New(),
		Position:                position.New(),
		Project:                 project.New(),
		ProjectCommissionConfig: projectcommissionconfig.New(),
		ProjectHead:             projecthead.New(),
		ProjectMember:           projectmember.New(),
		ProjectMemberPosition:   projectmemberposition.New(),
		ProjectNotion:           projectnotion.New(),
		ProjectSlot:             projectslot.New(),
		ProjectSlotPosition:     projectslotposition.New(),
		ProjectStack:            projectstack.New(),
		Question:                question.New(),
		Recruitment:             recruitment.New(),
		Role:                    role.New(),
		Schedule:                schedule.New(),
		Seniority:               seniority.New(),
		SocialAccount:           socialaccount.New(),
		Stack:                   stack.New(),
		Valuation:               valuation.New(),
		WorkUnit:                workunit.New(),
		WorkUnitMember:          workunitmember.New(),
		WorkUnitStack:           workunitstack.New(),
		OperationalService:      operationalservice.New(),
		DiscordLogTemplate:      discordtemplate.New(),
	}
}
