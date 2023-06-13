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
	PermissionAssetUpload                         PermissionCode = "assets.upload"
	PermissionAuthCreate                          PermissionCode = "auth.create"
	PermissionAuthRead                            PermissionCode = "auth.read"
	PermissionBankAccountRead                     PermissionCode = "bankAccounts.read"
	PermissionClientCreate                        PermissionCode = "clients.create"
	PermissionClientDelete                        PermissionCode = "clients.delete"
	PermissionClientEdit                          PermissionCode = "clients.edit"
	PermissionClientRead                          PermissionCode = "clients.read"
	PermissionCronjobExecute                      PermissionCode = "cronjobs.execute"
	PermissionDashBoardEngagementRead             PermissionCode = "dashboards.engagement.read"
	PermissionDashBoardProjectsRead               PermissionCode = "dashboards.projects.read"
	PermissionDashBoardRead                       PermissionCode = "dashboards.read"
	PermissionDashBoardResourcesRead              PermissionCode = "dashboards.resources.read"
	PermissionEarnRead                            PermissionCode = "earns.read"
	PermissionEmployeeEventQuestionsCreate        PermissionCode = "employeeEventQuestions.create"
	PermissionEmployeeEventQuestionsDelete        PermissionCode = "employeeEventQuestions.delete"
	PermissionEmployeeEventQuestionsEdit          PermissionCode = "employeeEventQuestions.edit"
	PermissionEmployeeEventQuestionsRead          PermissionCode = "employeeEventQuestions.read"
	PermissionEmployeeMenteesCreate               PermissionCode = "employeeMentees.create"
	PermissionEmployeeMenteesDelete               PermissionCode = "employeeMentees.delete"
	PermissionEmployeeMenteesEdit                 PermissionCode = "employeeMentees.edit"
	PermissionEmployeeMenteesRead                 PermissionCode = "employeeMentees.read"
	PermissionEmployeeRolesCreate                 PermissionCode = "employeeRoles.create"
	PermissionEmployeeRolesDelete                 PermissionCode = "employeeRoles.delete"
	PermissionEmployeeRolesEdit                   PermissionCode = "employeeRoles.edit"
	PermissionEmployeeRolesRead                   PermissionCode = "employeeRoles.read"
	PermissionEmployeesBaseSalaryEdit             PermissionCode = "employees.baseSalary.edit"
	PermissionEmployeesBaseSalaryRead             PermissionCode = "employees.baseSalary.read"
	PermissionEmployeesCreate                     PermissionCode = "employees.create"
	PermissionEmployeesDelete                     PermissionCode = "employees.delete"
	PermissionEmployeesEdit                       PermissionCode = "employees.edit"
	PermissionEmployeesRead                       PermissionCode = "employees.read"
	PermissionEmployeesReadFilterByAllStatuses    PermissionCode = "employees.read.filterByAllStatuses"
	PermissionEmployeesReadFilterByProject        PermissionCode = "employees.read.filterByProject"
	PermissionEmployeesReadFullAccess             PermissionCode = "employees.read.fullAccess"
	PermissionEmployeesReadGeneralInfoFullAccess  PermissionCode = "employees.read.generalInfo.fullAccess"
	PermissionEmployeesReadLineManagerFullAccess  PermissionCode = "employees.read.lineManager.fullAccess"
	PermissionEmployeesReadPersonalInfoFullAccess PermissionCode = "employees.read.personalInfo.fullAccess"
	PermissionEmployeesReadProjectsFullAccess     PermissionCode = "employees.read.projects.fullAccess"
	PermissionEmployeesReadProjectsReadActive     PermissionCode = "employees.read.projects.readActive"
	PermissionEmployeesReadReadActive             PermissionCode = "employees.read.readActive"
	PermissionFeedbacksCreate                     PermissionCode = "feedbacks.create"
	PermissionFeedbacksDelete                     PermissionCode = "feedbacks.delete"
	PermissionFeedbacksEdit                       PermissionCode = "feedbacks.edit"
	PermissionFeedbacksRead                       PermissionCode = "feedbacks.read"
	PermissionInvoiceCreate                       PermissionCode = "invoices.create"
	PermissionInvoiceDelete                       PermissionCode = "invoices.delete"
	PermissionInvoiceEdit                         PermissionCode = "invoices.edit"
	PermissionInvoiceRead                         PermissionCode = "invoices.read"
	PermissionMetadataCreate                      PermissionCode = "metadata.create"
	PermissionMetadataDelete                      PermissionCode = "metadata.delete"
	PermissionMetadataEdit                        PermissionCode = "metadata.edit"
	PermissionMetadataRead                        PermissionCode = "metadata.read"
	PermissionNotionCreate                        PermissionCode = "notion.create"
	PermissionNotionRead                          PermissionCode = "notion.read"
	PermissionNotionSend                          PermissionCode = "notion.send"
	PermissionPayrollsCreate                      PermissionCode = "payrolls.create"
	PermissionPayrollsEdit                        PermissionCode = "payrolls.edit"
	PermissionPayrollsRead                        PermissionCode = "payrolls.read"
	PermissionProjectMembersCreate                PermissionCode = "projectMembers.create"
	PermissionProjectMembersDelete                PermissionCode = "projectMembers.delete"
	PermissionProjectMembersEdit                  PermissionCode = "projectMembers.edit"
	PermissionProjectMembersRateEdit              PermissionCode = "projectMembers.rate.edit"
	PermissionProjectMembersRateRead              PermissionCode = "projectMembers.rate.read"
	PermissionProjectMembersRead                  PermissionCode = "projectMembers.read"
	PermissionProjectWorkUnitsCreate              PermissionCode = "projectWorkUnits.create"
	PermissionProjectWorkUnitsCreateFullAccess    PermissionCode = "projectWorkUnits.create.fullAccess"
	PermissionProjectWorkUnitsDelete              PermissionCode = "projectWorkUnits.delete"
	PermissionProjectWorkUnitsDeleteFullAccess    PermissionCode = "projectWorkUnits.delete.fullAccess"
	PermissionProjectWorkUnitsEdit                PermissionCode = "projectWorkUnits.edit"
	PermissionProjectWorkUnitsEditFullAccess      PermissionCode = "projectWorkUnits.edit.fullAccess"
	PermissionProjectWorkUnitsRead                PermissionCode = "projectWorkUnits.read"
	PermissionProjectWorkUnitsReadFullAccess      PermissionCode = "projectWorkUnits.read.fullAccess"
	PermissionProjectsCommissionRateEdit          PermissionCode = "projects.commissionRate.edit"
	PermissionProjectsCommissionRateRead          PermissionCode = "projects.commissionRate.read"
	PermissionProjectsCreate                      PermissionCode = "projects.create"
	PermissionProjectsEdit                        PermissionCode = "projects.edit"
	PermissionProjectsRead                        PermissionCode = "projects.read"
	PermissionProjectsReadFullAccess              PermissionCode = "projects.read.fullAccess"
	PermissionProjectsReadMonthlyRevenue          PermissionCode = "projects.read.monthlyRevenue"
	PermissionProjectsReadReadActive              PermissionCode = "projects.read.readActive"
	PermissionSurveysCreate                       PermissionCode = "surveys.create"
	PermissionSurveysDelete                       PermissionCode = "surveys.delete"
	PermissionSurveysEdit                         PermissionCode = "surveys.edit"
	PermissionSurveysRead                         PermissionCode = "surveys.read"
	PermissionValuationRead                       PermissionCode = "valuations.read"
	PermissionEngagementMetricsWrite              PermissionCode = "engagementMetrics.write"
	PermissionEngagementMetricsRead               PermissionCode = "engagementMetrics.read"
	PermissionIcyDistributionRead                 PermissionCode = "icyDistribution.read"
)

func (p PermissionCode) String() string {
	return string(p)
}
