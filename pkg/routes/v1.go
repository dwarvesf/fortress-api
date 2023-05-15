package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/mw"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

func loadV1Routes(r *gin.Engine, h *handler.Handler, repo store.DBRepo, s *store.Store, cfg *config.Config) {
	v1 := r.Group("/api/v1")
	cronjob := r.Group("/cronjobs")
	webhook := r.Group("/webhooks")

	pmw := mw.NewPermissionMiddleware(s, repo, cfg)
	amw := mw.NewAuthMiddleware(cfg, s, repo)

	// cronjob group
	{
		cronjob.POST("/audits", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Audit.Sync)
		cronjob.POST("/birthday", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Birthday.BirthdayDailyMessage)
		cronjob.POST("/sync-discord-info", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Discord.SyncDiscordInfo)
		cronjob.POST("/sync-monthly-accounting-todo", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Accounting.CreateAccountingTodo)
		cronjob.PUT("/sync-project-member-status", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Project.SyncProjectMemberStatus)
		cronjob.POST("/store-vault-transaction", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Vault.StoreVaultTransaction)
	}

	/////////////////
	// Webhook GROUP
	/////////////////
	webhook.POST("/n8n", h.Webhook.N8n)
	// Basecamp
	basecampGroup := webhook.Group("/basecamp")
	{
		expenseGroup := basecampGroup.Group("/expense")
		{
			expenseGroup.POST("/validate", h.Webhook.BasecampExpenseValidate)
			expenseGroup.POST("", h.Webhook.BasecampExpense)
			expenseGroup.DELETE("", h.Webhook.UncheckBasecampExpense)
		}
		operationGroup := basecampGroup.Group("/operation")
		{
			operationGroup.POST("/accounting-transaction", h.Webhook.StoreAccountingTransaction)
			operationGroup.PUT("/invoice", h.Webhook.MarkInvoiceAsPaidViaBasecamp)
		}
	}

	/////////////////
	// API GROUP
	/////////////////

	// auth
	v1.POST("/auth", h.Auth.Auth)
	v1.GET("/auth/me", amw.WithAuth, pmw.WithPerm(model.PermissionAuthRead), h.Auth.Me)
	v1.POST("/auth/api-key", amw.WithAuth, pmw.WithPerm(model.PermissionAuthCreate), h.Auth.CreateAPIKey)

	// user profile
	v1.GET("/profile", amw.WithAuth, h.Profile.GetProfile)
	v1.PUT("/profile", amw.WithAuth, h.Profile.UpdateInfo)
	v1.POST("/profile/upload-avatar", amw.WithAuth, h.Profile.UploadAvatar)

	// assets
	v1.POST("/assets/upload", amw.WithAuth, pmw.WithPerm(model.PermissionAssetUpload), h.Asset.Upload)

	// employees
	v1.POST("/employees", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesCreate), h.Employee.Create)
	v1.POST("/employees/search", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesRead), h.Employee.List)
	v1.GET("/employees/:id", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesRead), h.Employee.Details)
	v1.PUT("/employees/:id/general-info", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdateGeneralInfo)
	v1.PUT("/employees/:id/personal-info", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdatePersonalInfo)
	v1.PUT("/employees/:id/skills", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdateSkills)
	v1.PUT("/employees/:id/employee-status", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdateEmployeeStatus)
	v1.POST("/employees/:id/upload-avatar", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UploadAvatar)
	v1.PUT("/employees/:id/roles", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeeRolesEdit), h.Employee.UpdateRole)
	v1.PUT("/employees/:id/base-salary", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesBaseSalaryEdit), h.Employee.UpdateBaseSalary)

	v1.GET("/line-managers", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesRead), h.Employee.GetLineManagers)

	// metadata
	v1.GET("/metadata/working-status", h.Metadata.WorkingStatuses)
	v1.GET("/metadata/stacks", h.Metadata.Stacks)
	v1.GET("/metadata/seniorities", h.Metadata.Seniorities)
	v1.GET("/metadata/chapters", h.Metadata.Chapters)
	v1.GET("/metadata/organizations", h.Metadata.Organizations)
	v1.GET("/metadata/roles", h.Metadata.GetRoles)
	v1.GET("/metadata/positions", h.Metadata.Positions)
	v1.GET("/metadata/countries", h.Metadata.GetCountries)
	v1.GET("/metadata/currencies", h.Metadata.GetCurrencies)
	v1.GET("/metadata/countries/:country_id/cities", h.Metadata.GetCities)
	v1.GET("/metadata/project-statuses", h.Metadata.ProjectStatuses)
	v1.GET("/metadata/questions", h.Metadata.GetQuestions)
	v1.PUT("/metadata/stacks/:id", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataEdit), h.Metadata.UpdateStack)
	v1.POST("/metadata/stacks", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataCreate), h.Metadata.CreateStack)
	v1.DELETE("/metadata/stacks/:id", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataDelete), h.Metadata.DeleteStack)
	v1.PUT("/metadata/positions/:id", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataEdit), h.Metadata.UpdatePosition)
	v1.POST("/metadata/positions", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataCreate), h.Metadata.CreatePosition)
	v1.DELETE("/metadata/positions/:id", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataDelete), h.Metadata.DeletePosition)

	// projects
	projectGroup := v1.Group("projects")
	{
		projectGroup.POST("", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsCreate), h.Project.Create)
		projectGroup.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsRead), h.Project.List)
		projectGroup.GET("/:id", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsRead), h.Project.Details)
		projectGroup.PUT("/:id/sending-survey-state", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UpdateSendingSurveyState)
		projectGroup.POST("/:id/upload-avatar", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UploadAvatar)
		projectGroup.PUT("/:id/status", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UpdateProjectStatus)
		projectGroup.POST("/:id/members", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersCreate), h.Project.AssignMember)
		projectGroup.GET("/:id/members", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersRead), h.Project.GetMembers)
		projectGroup.PUT("/:id/members", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersEdit), h.Project.UpdateMember)
		// projectGroup.PUT("/:id/members/:memberID", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersEdit), h.Project.UnassignMember)
		projectGroup.DELETE("/:id/members/:memberID", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersDelete), h.Project.DeleteMember)
		projectGroup.DELETE("/:id/slots/:slotID", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersDelete), h.Project.DeleteSlot)
		projectGroup.PUT("/:id/general-info", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UpdateGeneralInfo)
		projectGroup.PUT("/:id/contact-info", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UpdateContactInfo)
		projectGroup.GET("/:id/work-units", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsRead), h.Project.GetWorkUnits)
		projectGroup.POST("/:id/work-units", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsCreate), h.Project.CreateWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsEdit), h.Project.UpdateWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID/archive", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsEdit), h.Project.ArchiveWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID/unarchive", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsEdit), h.Project.UnarchiveWorkUnit)
	}

	clientGroup := v1.Group("/clients")
	{
		clientGroup.POST("", amw.WithAuth, pmw.WithPerm(model.PermissionClientCreate), h.Client.Create)
		clientGroup.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionClientRead), h.Client.List)
		clientGroup.GET("/:id", amw.WithAuth, pmw.WithPerm(model.PermissionClientEdit), h.Client.Detail)
		clientGroup.PUT("/:id", amw.WithAuth, pmw.WithPerm(model.PermissionClientRead), h.Client.Update)
		clientGroup.DELETE("/:id", amw.WithAuth, pmw.WithPerm(model.PermissionClientDelete), h.Client.Delete)
	}

	feedbackGroup := v1.Group("/feedbacks")
	{
		feedbackGroup.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionFeedbacksRead), h.Feedback.List)
		feedbackGroup.GET("/:id/topics/:topicID", amw.WithAuth, pmw.WithPerm(model.PermissionFeedbacksRead), h.Feedback.Detail)
		feedbackGroup.POST("/:id/topics/:topicID/submit", amw.WithAuth, pmw.WithPerm(model.PermissionFeedbacksCreate), h.Feedback.Submit)
		feedbackGroup.GET("/unreads", amw.WithAuth, pmw.WithPerm(model.PermissionFeedbacksRead), h.Feedback.CountUnreadFeedback)
	}

	surveyGroup := v1.Group("/surveys")
	{
		surveyGroup.POST("", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysCreate), h.Survey.CreateSurvey)
		surveyGroup.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysRead), h.Survey.ListSurvey)
		surveyGroup.GET("/:id", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysRead), h.Survey.GetSurveyDetail)
		surveyGroup.DELETE("/:id", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysDelete), h.Survey.DeleteSurvey)
		surveyGroup.POST("/:id/send", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysCreate), h.Survey.SendSurvey)
		surveyGroup.GET("/:id/topics/:topicID/reviews/:reviewID", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeeEventQuestionsRead), h.Survey.GetSurveyReviewDetail)
		surveyGroup.DELETE("/:id/topics/:topicID", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysDelete), h.Survey.DeleteSurveyTopic)
		surveyGroup.GET("/:id/topics/:topicID", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysRead), h.Survey.GetSurveyTopicDetail)
		surveyGroup.PUT("/:id/topics/:topicID/employees", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysEdit), h.Survey.UpdateTopicReviewers)
		surveyGroup.PUT("/:id/done", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysEdit), h.Survey.MarkDone)
		surveyGroup.DELETE("/:id/topics/:topicID/employees", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysEdit), h.Survey.DeleteTopicReviewers)
	}

	bankGroup := v1.Group("/bank-accounts")
	{
		bankGroup.GET("", pmw.WithPerm(model.PermissionBankAccountRead), h.BankAccount.List)
	}

	invoiceGroup := v1.Group("/invoices")
	{
		invoiceGroup.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionInvoiceRead), h.Invoice.List)
		invoiceGroup.PUT("/:id/status", amw.WithAuth, pmw.WithPerm(model.PermissionInvoiceEdit), h.Invoice.UpdateStatus)
		invoiceGroup.GET("/template", amw.WithAuth, pmw.WithPerm(model.PermissionInvoiceRead), h.Invoice.GetTemplate)
		invoiceGroup.POST("/send", amw.WithAuth, pmw.WithPerm(model.PermissionInvoiceRead), h.Invoice.Send)
	}

	valuation := v1.Group("/valuation")
	{
		valuation.GET("/:year", pmw.WithPerm(model.PermissionValuationRead), h.Valuation.One)
	}

	notion := v1.Group("/notion")
	{
		earn := notion.Group("/earn")
		{
			earn.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.ListEarns)
		}
		techRadar := notion.Group("/tech-radar")
		{
			techRadar.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.ListTechRadars)
			techRadar.POST("", amw.WithAuth, pmw.WithPerm(model.PermissionNotionCreate), h.Notion.CreateTechRadar)
		}
		audience := notion.Group("/audiences")
		{
			audience.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.ListAudiences)
		}
		event := notion.Group("/events")
		{
			event.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.ListEvents)
		}
		digest := notion.Group("/digests")
		{
			digest.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.ListDigests)
		}
		update := notion.Group("/updates")
		{
			update.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.ListUpdates)
		}
		memo := notion.Group("/memos")
		{
			memo.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.ListMemos)
		}
		issue := notion.Group("/issues")
		{
			issue.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.ListIssues)
		}
		staffingDemand := notion.Group("/staffing-demands")
		{
			staffingDemand.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.ListStaffingDemands)
		}
		hiring := notion.Group("/hiring-positions")
		{
			hiring.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.ListHiringPositions)
		}

		projectNotion := notion.Group("/projects")
		{
			projectNotion.GET("/milestones", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.ListProjectMilestones)
		}

		dfUpdates := notion.Group("df-updates")
		{
			dfUpdates.POST("/:id/send", amw.WithAuth, pmw.WithPerm(model.PermissionNotionSend), h.Notion.SendNewsLetter)
		}

		notionChangelog := notion.Group("changelogs")
		{
			notionChangelog.GET("/projects/available", amw.WithAuth, pmw.WithPerm(model.PermissionNotionRead), h.Notion.GetAvailableProjectsChangelog)
			notionChangelog.POST("/project", amw.WithAuth, pmw.WithPerm(model.PermissionNotionSend), h.Notion.SendProjectChangelog)
		}
	}

	dashboard := v1.Group("/dashboards")
	{
		engagementDashboardGroup := dashboard.Group("/engagement", amw.WithAuth, pmw.WithPerm(model.PermissionDashBoardEngagementRead))
		{
			engagementDashboardGroup.GET("/info", h.Dashboard.GetEngagementInfo)
			engagementDashboardGroup.GET("/detail", h.Dashboard.GetEngagementInfoDetail)
		}

		projectDashboardGroup := dashboard.Group("/projects", amw.WithAuth, pmw.WithPerm(model.PermissionDashBoardProjectsRead))
		{
			projectDashboardGroup.GET("/sizes", h.Dashboard.GetProjectSizes)
			projectDashboardGroup.GET("/work-surveys", h.Dashboard.GetWorkSurveys)
			projectDashboardGroup.GET("/action-items", h.Dashboard.GetActionItemReports)
			projectDashboardGroup.GET("/engineering-healths", h.Dashboard.GetEngineeringHealth)
			projectDashboardGroup.GET("/audits", h.Dashboard.GetAudits)
			projectDashboardGroup.GET("/action-item-squash", h.Dashboard.GetActionItemSquashReports)
			projectDashboardGroup.GET("/summary", h.Dashboard.GetSummary)
		}

		resourceDashboardGroup := dashboard.Group("/resources", amw.WithAuth, pmw.WithPerm(model.PermissionDashBoardResourcesRead))
		{
			resourceDashboardGroup.GET("/availabilities", h.Dashboard.GetResourcesAvailability)
			resourceDashboardGroup.GET("/utilization", h.Dashboard.GetResourceUtilization)
			resourceDashboardGroup.GET("/work-unit-distribution", h.Dashboard.GetWorkUnitDistribution)
			resourceDashboardGroup.GET("/work-unit-distribution-summary", h.Dashboard.GetWorkUnitDistributionSummary)
			resourceDashboardGroup.GET("/work-survey-summaries", h.Dashboard.GetResourceWorkSurveySummaries)
		}
	}

	payroll := v1.Group("payrolls")
	{
		payroll.PUT("", amw.WithAuth, pmw.WithPerm(model.PermissionPayrollsEdit), h.Payroll.MarkPayrollAsPaid)
		payroll.GET("/detail", amw.WithAuth, pmw.WithPerm(model.PermissionPayrollsRead), h.Payroll.GetPayrollsByMonth)
		payroll.GET("/bhxh", amw.WithAuth, pmw.WithPerm(model.PermissionPayrollsRead), h.Payroll.GetPayrollsBHXH)
		payroll.POST("/commit", amw.WithAuth, pmw.WithPerm(model.PermissionPayrollsCreate), h.Payroll.CommitPayroll)
	}
}
