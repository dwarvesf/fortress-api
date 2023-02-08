package model

type Permission struct {
	BaseModel

	Code string
	Name string
}

func ToPermissionMap(perms []*Permission) map[string]string {
	m := make(map[string]string)
	for _, v := range perms {
		m[v.Code] = v.Name
	}

	return m
}

type PermissionCode string

const (
	PermissionAuthRead                            PermissionCode = "auth.read"
	PermissionEmployeesRead                       PermissionCode = "employees.read"
	PermissionEmployeesReadFullAccess             PermissionCode = "employees.read.fullAccess"
	PermissionEmployeesReadReadActive             PermissionCode = "employees.read.readActive"
	PermissionEmployeesReadGeneralInfoFullAccess  PermissionCode = "employees.read.generalInfo.fullAccess"
	PermissionEmployeesReadPersonalInfoFullAccess PermissionCode = "employees.read.personalInfo.fullAccess"
	PermissionEmployeesReadProjectsFullAccess     PermissionCode = "employees.read.projects.fullAccess"
	PermissionEmployeesReadProjectsReadActive     PermissionCode = "employees.read.projects.readActive"
	PermissionEmployeesReadFilterByAllStatuses    PermissionCode = "employees.read.filterByAllStatuses"
	PermissionEmployeesReadFilterByProject        PermissionCode = "employees.read.filterByProject"
	PermissionEmployeesReadLineManagerFullAccess  PermissionCode = "employees.read.lineManager.fullAccess"
	PermissionEmployeesCreate                     PermissionCode = "employees.create"
	PermissionEmployeesEdit                       PermissionCode = "employees.edit"
	PermissionEmployeesDelete                     PermissionCode = "employees.delete"
	PermissionProjectsCreate                      PermissionCode = "projects.create"
	PermissionProjectsRead                        PermissionCode = "projects.read"
	PermissionProjectsReadFullAccess              PermissionCode = "projects.read.fullAccess"
	PermissionProjectsReadReadActive              PermissionCode = "projects.read.readActive"
	PermissionProjectsEdit                        PermissionCode = "projects.edit"
	PermissionProjectMembersCreate                PermissionCode = "projectMembers.create"
	PermissionProjectMembersRead                  PermissionCode = "projectMembers.read"
	PermissionProjectMembersEdit                  PermissionCode = "projectMembers.edit"
	PermissionProjectMembersDelete                PermissionCode = "projectMembers.delete"
	PermissionProjectWorkUnitsCreate              PermissionCode = "projectWorkUnits.create"
	PermissionProjectWorkUnitsCreateFullAccess    PermissionCode = "projectWorkUnits.create.fullAccess"
	PermissionProjectWorkUnitsRead                PermissionCode = "projectWorkUnits.read"
	PermissionProjectWorkUnitsReadFullAccess      PermissionCode = "projectWorkUnits.read.fullAccess"
	PermissionProjectWorkUnitsEdit                PermissionCode = "projectWorkUnits.edit"
	PermissionProjectWorkUnitsEditFullAccess      PermissionCode = "projectWorkUnits.edit.fullAccess"
	PermissionProjectWorkUnitsDelete              PermissionCode = "projectWorkUnits.delete"
	PermissionProjectWorkUnitsDeleteFullAccess    PermissionCode = "projectWorkUnits.delete.fullAccess"
	PermissionFeedbacksCreate                     PermissionCode = "feedbacks.create"
	PermissionFeedbacksRead                       PermissionCode = "feedbacks.read"
	PermissionFeedbacksEdit                       PermissionCode = "feedbacks.edit"
	PermissionFeedbacksDelete                     PermissionCode = "feedbacks.delete"
	PermissionEmployeeEventQuestionsCreate        PermissionCode = "employeeEventQuestions.create"
	PermissionEmployeeEventQuestionsEdit          PermissionCode = "employeeEventQuestions.edit"
	PermissionEmployeeEventQuestionsRead          PermissionCode = "employeeEventQuestions.read"
	PermissionEmployeeEventQuestionsDelete        PermissionCode = "employeeEventQuestions.delete"
	PermissionSurveysCreate                       PermissionCode = "surveys.create"
	PermissionSurveysRead                         PermissionCode = "surveys.read"
	PermissionSurveysEdit                         PermissionCode = "surveys.edit"
	PermissionSurveysDelete                       PermissionCode = "surveys.delete"
	PermissionEmployeeMenteesCreate               PermissionCode = "employeeMentees.create"
	PermissionEmployeeMenteesRead                 PermissionCode = "employeeMentees.read"
	PermissionEmployeeMenteesEdit                 PermissionCode = "employeeMentees.edit"
	PermissionEmployeeMenteesDelete               PermissionCode = "employeeMentees.delete"
	PermissionMetadataCreate                      PermissionCode = "metadata.create"
	PermissionMetadataRead                        PermissionCode = "metadata.read"
	PermissionMetadataEdit                        PermissionCode = "metadata.edit"
	PermissionMetadataDelete                      PermissionCode = "metadata.delete"
	PermissionEmployeeRolesCreate                 PermissionCode = "employeeRoles.create"
	PermissionEmployeeRolesRead                   PermissionCode = "employeeRoles.read"
	PermissionEmployeeRolesEdit                   PermissionCode = "employeeRoles.edit"
	PermissionEmployeeRolesDelete                 PermissionCode = "employeeRoles.delete"
	PermissionValuationRead                       PermissionCode = "valuations.read"
	PermissionEarnRead                            PermissionCode = "earns.read"
	PermissionInvoiceCreate                       PermissionCode = "invoices.create"
	PermissionInvoiceRead                         PermissionCode = "invoices.read"
	PermissionInvoiceEdit                         PermissionCode = "invoices.edit"
	PermissionInvoiceDelete                       PermissionCode = "invoices.delete"
)

func (p PermissionCode) String() string {
	return string(p)
}
