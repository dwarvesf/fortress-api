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
	// if r == nil || h == nil || repo == nil || s == nil || cfg == nil {
	// 	return
	// }

	pmw := mw.NewPermissionMiddleware(s, repo, cfg)
	amw := mw.NewAuthMiddleware(cfg, s, repo)

	// Define conditional middleware based on environment
	var conditionalAuthMW gin.HandlerFunc
	var conditionalPermMW func(model.PermissionCode) gin.HandlerFunc

	if cfg.Env == "local" {
		// Bypass middleware in local environment
		conditionalAuthMW = func(c *gin.Context) { c.Next() }
		conditionalPermMW = func(_ model.PermissionCode) gin.HandlerFunc { return func(c *gin.Context) { c.Next() } }
	} else {
		// Use actual middleware in other environments
		conditionalAuthMW = amw.WithAuth
		conditionalPermMW = func(perm model.PermissionCode) gin.HandlerFunc {
			return pmw.WithPerm(perm)
		}
	}

	/////////////////
	// Cronjob GROUP
	/////////////////
	cronjob := r.Group("/cronjobs")
	{
		cronjob.POST("/audits", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Audit.Sync)
		cronjob.POST("/birthday", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Discord.BirthdayDailyMessage)
		cronjob.POST("/on-leaves", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Discord.OnLeaveMessage)
		cronjob.POST("/sync-discord-info", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Discord.SyncDiscordInfo)
		cronjob.POST("/sync-monthly-accounting-todo", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Accounting.CreateAccountingTodo)
		cronjob.POST("/sync-project-member-status", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Project.SyncProjectMemberStatus)
		cronjob.POST("/store-vault-transaction", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Vault.StoreVaultTransaction)
		cronjob.POST("/index-engagement-messages", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Engagement.IndexMessages)
		cronjob.POST("/brainery-reports", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Discord.ReportBraineryMetrics)
		cronjob.POST("/delivery-metric-reports", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Discord.DeliveryMetricsReport)
		cronjob.POST("/sync-delivery-metrics", conditionalAuthMW, conditionalPermMW(model.PermissionDeliveryMetricsSync), h.DeliveryMetric.Sync)
		cronjob.POST("/sync-conversion-rates", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.ConversionRate.Sync)
		cronjob.POST("/sync-memo", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Discord.SyncMemo)
		cronjob.POST("/sweep-memo", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Discord.SweepMemo)
		cronjob.POST("/notify-weekly-memos", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Discord.NotifyWeeklyMemos)
		cronjob.POST("/notify-top-memo-authors", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Discord.NotifyTopMemoAuthors)
		cronjob.POST("/transcribe-youtube-broadcast", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Youtube.TranscribeBroadcast)
		cronjob.POST("/sweep-ogif-event", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Discord.SweepOgifEvent)
		cronjob.POST("/sync-project-heads", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Project.SyncProjectHeadsFromNotion)
	}

	/////////////////
	// Webhook GROUP
	/////////////////
	webhook := r.Group("/webhooks")
	{
		webhook.POST("/n8n", h.Webhook.N8n)

		basecampGroup := webhook.Group("/basecamp")
		{
			expenseGroup := basecampGroup.Group("/expense")
			{
				expenseGroup.POST("/validate", h.Webhook.ValidateBasecampExpense)
				expenseGroup.POST("", h.Webhook.CreateBasecampExpense)
				expenseGroup.DELETE("", h.Webhook.UncheckBasecampExpense)
			}
			operationGroup := basecampGroup.Group("/operation")
			{
				operationGroup.POST("/accounting-transaction", h.Webhook.StoreAccountingTransaction)
				operationGroup.PUT("/invoice", h.Webhook.MarkInvoiceAsPaidViaBasecamp)
			}
			onLeaveGroup := basecampGroup.Group("/onleave")
			{
				onLeaveGroup.POST("/validate", h.Webhook.ValidateOnLeaveRequest)
				onLeaveGroup.POST("", h.Webhook.ApproveOnLeaveRequest)
			}
		}
	}

	/////////////////
	// API GROUP
	/////////////////
	v1 := r.Group("/api/v1")

	// assets
	assetGroup := v1.Group("/assets")
	{
		assetGroup.POST("/upload", conditionalAuthMW, h.Asset.Upload)
	}

	lineManagerGroup := v1.Group("/line-managers")
	{
		lineManagerGroup.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesRead), h.Employee.GetLineManagers)
	}

	// auth
	authRoute := v1.Group("/auth")
	{
		authRoute.POST("", h.Auth.Auth)
		authRoute.GET("/me", conditionalAuthMW, conditionalPermMW(model.PermissionAuthRead), h.Auth.Me)
		authRoute.POST("/api-key", conditionalAuthMW, conditionalPermMW(model.PermissionAuthCreate), h.Auth.CreateAPIKey)
	}

	// user profile
	profileGroup := v1.Group("/profile")
	{
		profileGroup.GET("", conditionalAuthMW, h.Profile.GetProfile)
		profileGroup.PUT("", conditionalAuthMW, h.Profile.UpdateInfo)
		profileGroup.POST("/upload-avatar", conditionalAuthMW, h.Profile.UploadAvatar)
		profileGroup.POST("/upload", conditionalAuthMW, h.Profile.Upload)
	}

	// employees
	employeeRoute := v1.Group("/employees")
	{
		employeeRoute.POST("", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesCreate), h.Employee.Create)
		employeeRoute.POST("/search", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesRead), h.Employee.List)
		employeeRoute.GET("/:id", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesRead), h.Employee.Details)
		employeeRoute.PUT("/:id/general-info", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesEdit), h.Employee.UpdateGeneralInfo)
		employeeRoute.PUT("/:id/personal-info", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesEdit), h.Employee.UpdatePersonalInfo)
		employeeRoute.PUT("/:id/skills", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesEdit), h.Employee.UpdateSkills)
		employeeRoute.PUT("/:id/employee-status", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesEdit), h.Employee.UpdateEmployeeStatus)
		employeeRoute.POST("/:id/upload-avatar", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesEdit), h.Employee.UploadAvatar)
		employeeRoute.PUT("/:id/roles", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeeRolesEdit), h.Employee.UpdateRole)
		employeeRoute.PUT("/:id/base-salary", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesBaseSalaryEdit), h.Employee.UpdateBaseSalary)
	}

	// metadata
	metadataRoute := v1.Group("/metadata")
	{
		metadataRoute.GET("/working-status", h.Metadata.WorkingStatuses)
		metadataRoute.GET("/banks", h.Metadata.Banks)
		metadataRoute.GET("/stacks", h.Metadata.Stacks)
		metadataRoute.GET("/seniorities", h.Metadata.Seniorities)
		metadataRoute.GET("/chapters", h.Metadata.Chapters)
		metadataRoute.GET("/organizations", h.Metadata.Organizations)
		metadataRoute.GET("/roles", h.Metadata.GetRoles)
		metadataRoute.GET("/positions", h.Metadata.Positions)
		metadataRoute.GET("/countries", h.Metadata.GetCountries)
		metadataRoute.GET("/currencies", h.Metadata.GetCurrencies)
		metadataRoute.GET("/countries/:country_id/cities", h.Metadata.GetCities)
		metadataRoute.GET("/project-statuses", h.Metadata.ProjectStatuses)
		metadataRoute.GET("/questions", h.Metadata.GetQuestions)
		metadataRoute.PUT("/stacks/:id", conditionalAuthMW, conditionalPermMW(model.PermissionMetadataEdit), h.Metadata.UpdateStack)
		metadataRoute.POST("/stacks", conditionalAuthMW, conditionalPermMW(model.PermissionMetadataCreate), h.Metadata.CreateStack)
		metadataRoute.DELETE("/stacks/:id", conditionalAuthMW, conditionalPermMW(model.PermissionMetadataDelete), h.Metadata.DeleteStack)
		metadataRoute.PUT("/positions/:id", conditionalAuthMW, conditionalPermMW(model.PermissionMetadataEdit), h.Metadata.UpdatePosition)
		metadataRoute.POST("/positions", conditionalAuthMW, conditionalPermMW(model.PermissionMetadataCreate), h.Metadata.CreatePosition)
		metadataRoute.DELETE("/positions/:id", conditionalAuthMW, conditionalPermMW(model.PermissionMetadataDelete), h.Metadata.DeletePosition)
	}

	// projects
	projectGroup := v1.Group("projects")
	{
		projectGroup.POST("", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsCreate), h.Project.Create)
		projectGroup.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsRead), h.Project.List)
		projectGroup.GET("/:id", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsRead), h.Project.Details)
		projectGroup.PUT("/:id/sending-survey-state", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsEdit), h.Project.UpdateSendingSurveyState)
		projectGroup.POST("/:id/upload-avatar", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsEdit), h.Project.UploadAvatar)
		projectGroup.PUT("/:id/status", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsEdit), h.Project.UpdateProjectStatus)
		projectGroup.POST("/:id/members", conditionalAuthMW, conditionalPermMW(model.PermissionProjectMembersCreate), h.Project.AssignMember)
		projectGroup.GET("/:id/members", conditionalAuthMW, conditionalPermMW(model.PermissionProjectMembersRead), h.Project.GetMembers)
		projectGroup.PUT("/:id/members", conditionalAuthMW, conditionalPermMW(model.PermissionProjectMembersEdit), h.Project.UpdateMember)
		projectGroup.DELETE("/:id/members/:memberID", conditionalAuthMW, conditionalPermMW(model.PermissionProjectMembersDelete), h.Project.DeleteMember)
		projectGroup.DELETE("/:id/slots/:slotID", conditionalAuthMW, conditionalPermMW(model.PermissionProjectMembersDelete), h.Project.DeleteSlot)
		projectGroup.PUT("/:id/general-info", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsEdit), h.Project.UpdateGeneralInfo)
		projectGroup.PUT("/:id/contact-info", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsEdit), h.Project.UpdateContactInfo)
		projectGroup.GET("/:id/work-units", conditionalAuthMW, conditionalPermMW(model.PermissionProjectWorkUnitsRead), h.Project.GetWorkUnits)
		projectGroup.POST("/:id/work-units", conditionalAuthMW, conditionalPermMW(model.PermissionProjectWorkUnitsCreate), h.Project.CreateWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID", conditionalAuthMW, conditionalPermMW(model.PermissionProjectWorkUnitsEdit), h.Project.UpdateWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID/archive", conditionalAuthMW, conditionalPermMW(model.PermissionProjectWorkUnitsEdit), h.Project.ArchiveWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID/unarchive", conditionalAuthMW, conditionalPermMW(model.PermissionProjectWorkUnitsEdit), h.Project.UnarchiveWorkUnit)
		projectGroup.GET("/:id/commission-models", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsCommissionModelsRead), h.Project.CommissionModels)
		projectGroup.GET("/icy-distribution/weekly", conditionalAuthMW, conditionalPermMW(model.PermissionIcyDistributionRead), h.Project.IcyWeeklyDistribution)
	}

	clientGroup := v1.Group("/clients")
	{
		clientGroup.POST("", conditionalAuthMW, conditionalPermMW(model.PermissionClientCreate), h.Client.Create)
		clientGroup.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionClientRead), h.Client.List)
		clientGroup.GET("/:id", conditionalAuthMW, conditionalPermMW(model.PermissionClientEdit), h.Client.Detail)
		clientGroup.PUT("/:id", conditionalAuthMW, conditionalPermMW(model.PermissionClientRead), h.Client.Update)
		clientGroup.DELETE("/:id", conditionalAuthMW, conditionalPermMW(model.PermissionClientDelete), h.Client.Delete)
	}

	feedbackGroup := v1.Group("/feedbacks")
	{
		feedbackGroup.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionFeedbacksRead), h.Feedback.List)
		feedbackGroup.GET("/:id/topics/:topicID", conditionalAuthMW, conditionalPermMW(model.PermissionFeedbacksRead), h.Feedback.Detail)
		feedbackGroup.POST("/:id/topics/:topicID/submit", conditionalAuthMW, conditionalPermMW(model.PermissionFeedbacksCreate), h.Feedback.Submit)
		feedbackGroup.GET("/unreads", conditionalAuthMW, conditionalPermMW(model.PermissionFeedbacksRead), h.Feedback.CountUnreadFeedback)
	}

	surveyGroup := v1.Group("/surveys")
	{
		surveyGroup.POST("", conditionalAuthMW, conditionalPermMW(model.PermissionSurveysCreate), h.Survey.CreateSurvey)
		surveyGroup.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionSurveysRead), h.Survey.ListSurvey)
		surveyGroup.GET("/:id", conditionalAuthMW, conditionalPermMW(model.PermissionSurveysRead), h.Survey.GetSurveyDetail)
		surveyGroup.DELETE("/:id", conditionalAuthMW, conditionalPermMW(model.PermissionSurveysDelete), h.Survey.DeleteSurvey)
		surveyGroup.POST("/:id/send", conditionalAuthMW, conditionalPermMW(model.PermissionSurveysCreate), h.Survey.SendSurvey)
		surveyGroup.GET("/:id/topics/:topicID/reviews/:reviewID", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeeEventQuestionsRead), h.Survey.GetSurveyReviewDetail)
		surveyGroup.DELETE("/:id/topics/:topicID", conditionalAuthMW, conditionalPermMW(model.PermissionSurveysDelete), h.Survey.DeleteSurveyTopic)
		surveyGroup.GET("/:id/topics/:topicID", conditionalAuthMW, conditionalPermMW(model.PermissionSurveysRead), h.Survey.GetSurveyTopicDetail)
		surveyGroup.PUT("/:id/topics/:topicID/employees", conditionalAuthMW, conditionalPermMW(model.PermissionSurveysEdit), h.Survey.UpdateTopicReviewers)
		surveyGroup.PUT("/:id/done", conditionalAuthMW, conditionalPermMW(model.PermissionSurveysEdit), h.Survey.MarkDone)
		surveyGroup.DELETE("/:id/topics/:topicID/employees", conditionalAuthMW, conditionalPermMW(model.PermissionSurveysEdit), h.Survey.DeleteTopicReviewers)
	}

	bankGroup := v1.Group("/bank-accounts")
	{
		bankGroup.GET("", conditionalPermMW(model.PermissionBankAccountRead), h.BankAccount.List)
	}

	companyInfoGroup := v1.Group("/company-infos")
	{
		companyInfoGroup.GET("", conditionalPermMW(model.PermissionCompanyInfoRead), h.CompanyInfo.List)
	}

	invoiceGroup := v1.Group("/invoices")
	{
		invoiceGroup.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceRead), h.Invoice.List)
		invoiceGroup.PUT("/:id/status", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceEdit), h.Invoice.UpdateStatus)
		invoiceGroup.POST("/:id/calculate-commissions", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsCommissionRateEdit), h.Invoice.CalculateCommissions)
		invoiceGroup.GET("/template", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceRead), h.Invoice.GetTemplate)
		invoiceGroup.POST("/send", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceRead), h.Invoice.Send)
	}

	valuation := v1.Group("/valuation")
	{
		valuation.GET("/:year", conditionalPermMW(model.PermissionValuationRead), h.Valuation.One)
	}

	notion := v1.Group("/notion")
	{
		earn := notion.Group("/earn")
		{
			earn.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.ListEarns)
		}
		techRadar := notion.Group("/tech-radar")
		{
			techRadar.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.ListTechRadars)
			techRadar.POST("", conditionalAuthMW, conditionalPermMW(model.PermissionNotionCreate), h.Notion.CreateTechRadar)
		}
		audience := notion.Group("/audiences")
		{
			audience.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.ListAudiences)
		}
		event := notion.Group("/events")
		{
			event.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.ListEvents)
		}
		digest := notion.Group("/digests")
		{
			digest.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.ListDigests)
		}
		update := notion.Group("/updates")
		{
			update.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.ListUpdates)
		}
		memo := notion.Group("/memos")
		{
			memo.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.ListMemos)
		}
		issue := notion.Group("/issues")
		{
			issue.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.ListIssues)
		}
		staffingDemand := notion.Group("/staffing-demands")
		{
			staffingDemand.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.ListStaffingDemands)
		}
		hiring := notion.Group("/hiring-positions")
		{
			hiring.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.ListHiringPositions)
		}

		projectNotion := notion.Group("/projects")
		{
			projectNotion.GET("/milestones", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.ListProjectMilestones)
		}

		dfUpdates := notion.Group("df-updates")
		{
			dfUpdates.POST("/:id/send", conditionalAuthMW, conditionalPermMW(model.PermissionNotionSend), h.Notion.SendNewsLetter)
		}

		notionChangelog := notion.Group("changelogs")
		{
			notionChangelog.GET("/projects/available", conditionalAuthMW, conditionalPermMW(model.PermissionNotionRead), h.Notion.GetAvailableProjectsChangelog)
			notionChangelog.POST("/project", conditionalAuthMW, conditionalPermMW(model.PermissionNotionSend), h.Notion.SendProjectChangelog)
		}
	}

	dashboard := v1.Group("/dashboards")
	{
		engagementDashboardGroup := dashboard.Group("/engagement", conditionalAuthMW, conditionalPermMW(model.PermissionDashBoardEngagementRead))
		{
			engagementDashboardGroup.GET("/info", h.Dashboard.GetEngagementInfo)
			engagementDashboardGroup.GET("/detail", h.Dashboard.GetEngagementInfoDetail)
		}

		projectDashboardGroup := dashboard.Group("/projects", conditionalAuthMW, conditionalPermMW(model.PermissionDashBoardProjectsRead))
		{
			projectDashboardGroup.GET("/sizes", h.Dashboard.GetProjectSizes)
			projectDashboardGroup.GET("/work-surveys", h.Dashboard.GetWorkSurveys)
			projectDashboardGroup.GET("/action-items", h.Dashboard.GetActionItemReports)
			projectDashboardGroup.GET("/engineering-healths", h.Dashboard.GetEngineeringHealth)
			projectDashboardGroup.GET("/audits", h.Dashboard.GetAudits)
			projectDashboardGroup.GET("/action-item-squash", h.Dashboard.GetActionItemSquashReports)
			projectDashboardGroup.GET("/summary", h.Dashboard.GetSummary)
		}

		resourceDashboardGroup := dashboard.Group("/resources", conditionalAuthMW, conditionalPermMW(model.PermissionDashBoardResourcesRead))
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
		payroll.PUT("", conditionalAuthMW, conditionalPermMW(model.PermissionPayrollsEdit), h.Payroll.MarkPayrollAsPaid)
		payroll.GET("/detail", conditionalAuthMW, conditionalPermMW(model.PermissionPayrollsRead), h.Payroll.GetPayrollsByMonth)
		payroll.GET("/bhxh", conditionalAuthMW, conditionalPermMW(model.PermissionPayrollsRead), h.Payroll.GetPayrollsBHXH)
		payroll.POST("/commit", conditionalAuthMW, conditionalPermMW(model.PermissionPayrollsCreate), h.Payroll.CommitPayroll)
	}

	invitationGroup := v1.Group("/invite")
	{
		invitationGroup.GET("", conditionalAuthMW, h.Profile.GetInvitation)
		invitationGroup.PUT("/submit", conditionalAuthMW, h.Profile.SubmitOnboardingForm)
	}

	engagementsGroup := v1.Group("/engagements")
	{
		engagementsGroup.POST(
			"/rollup",
			conditionalAuthMW,
			conditionalPermMW(model.PermissionEngagementMetricsWrite),
			h.Engagement.UpsertRollup,
		)
		engagementsGroup.GET(
			"/channels/:channel-id/last-message-id",
			conditionalAuthMW,
			conditionalPermMW(model.PermissionEngagementMetricsRead),
			h.Engagement.GetLastMessageID,
		)
	}

	braineryGroup := v1.Group("/brainery-logs")
	{
		braineryGroup.POST("", conditionalAuthMW, conditionalPermMW(model.PermissionBraineryLogsWrite), h.BraineryLog.Create)
		braineryGroup.GET("/metrics", conditionalAuthMW, conditionalPermMW(model.PermissionBraineryLogsRead), h.BraineryLog.GetMetrics)
		braineryGroup.POST("/sync", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.BraineryLog.Sync)
	}

	memoGroup := v1.Group("/memos")
	{
		memoGroup.POST("", conditionalAuthMW, h.MemoLog.Create)
		memoGroup.POST("/sync", conditionalAuthMW, h.MemoLog.Sync)
		memoGroup.GET("", conditionalAuthMW, h.MemoLog.List)
		memoGroup.GET("/discords", conditionalAuthMW, h.MemoLog.ListByDiscordID)
		memoGroup.GET("/prs", conditionalAuthMW, h.MemoLog.ListOpenPullRequest)
		memoGroup.GET("/top-authors", conditionalAuthMW, h.MemoLog.GetTopAuthors)
	}

	earnGroup := v1.Group("/earns")
	{
		earnGroup.GET("", conditionalAuthMW, h.Earn.ListEarn)
	}

	// Delivery metrics
	{
		deliveryGroup := v1.Group("/delivery-metrics")
		deliveryGroup.POST("/report/sync", conditionalAuthMW, conditionalPermMW(model.PermissionDeliveryMetricsSync), h.DeliveryMetric.Sync)

		deliveryGroup.GET("/report/weekly", conditionalAuthMW, conditionalPermMW(model.PermissionDeliveryMetricsRead), h.DeliveryMetric.GetWeeklyReport)
		deliveryGroup.GET("/report/monthly", conditionalAuthMW, conditionalPermMW(model.PermissionDeliveryMetricsRead), h.DeliveryMetric.GetMonthlyReport)
		deliveryGroup.GET("/leader-board/weekly", conditionalAuthMW, conditionalPermMW(model.PermissionDeliveryMetricsLeaderBoardRead), h.DeliveryMetric.GetWeeklyLeaderBoard)
		deliveryGroup.GET("/leader-board/monthly", conditionalAuthMW, conditionalPermMW(model.PermissionDeliveryMetricsLeaderBoardRead), h.DeliveryMetric.GetMonthlyLeaderBoard)

		// API for fortress-discord
		deliveryGroup.GET("/report/weekly/discord-msg", conditionalAuthMW, conditionalPermMW(model.PermissionDeliveryMetricsRead), h.DeliveryMetric.GetWeeklyReportDiscordMsg)
		deliveryGroup.GET("/report/monthly/discord-msg", conditionalAuthMW, conditionalPermMW(model.PermissionDeliveryMetricsRead), h.DeliveryMetric.GetMonthlyReportDiscordMsg)
	}

	discordGroup := v1.Group("/discords")
	{
		discordGroup.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Employee.ListByDiscordRequest)
		discordGroup.GET("/mma-scores", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Employee.ListWithMMAScore)
		discordGroup.POST("/advance-salary", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Employee.SalaryAdvance)
		discordGroup.POST("/check-advance-salary", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Employee.CheckSalaryAdvance)

		discordGroup.GET("/salary-advance-report", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Employee.SalaryAdvanceReport)
		discordGroup.GET("/:discord_id/earns/transactions", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Employee.GetEmployeeEarnTransactions)
		discordGroup.GET("/:discord_id/earns/total", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Employee.GetEmployeeTotalEarn)
		discordGroup.GET("/earns/total", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Employee.GetTotalEarn)
		discordGroup.POST("/office-checkin", conditionalAuthMW, conditionalPermMW(model.PermissionTransferCheckinIcy), h.Employee.OfficeCheckIn)

		discordGroup.GET("/icy-accounting", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Icy.Accounting)

		scheduledEventGroup := discordGroup.Group("/scheduled-events")
		{
			scheduledEventGroup.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Discord.ListScheduledEvent)
			scheduledEventGroup.POST("", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordCreate), h.Discord.CreateScheduledEvent)
			scheduledEventGroup.PUT("/:id/speakers", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordEdit), h.Discord.SetScheduledEventSpeakers)
		}

		discordGroup.GET("/research-topics", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Discord.ListDiscordResearchTopics)
	}

	conversionRateGroup := v1.Group("/conversion-rates")
	{
		conversionRateGroup.GET("", conditionalAuthMW, h.ConversionRate.List)
	}

	newsGroup := v1.Group("/news")
	{
		newsGroup.GET("", conditionalAuthMW, h.News.Fetch)
	}

	ogifGroup := v1.Group("/ogif")
	{
		ogifGroup.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Discord.UserOgifStats)
		ogifGroup.GET("/leaderboard", conditionalAuthMW, conditionalPermMW(model.PermissionEmployeesDiscordRead), h.Discord.OgifLeaderboard)
	}

	// dynamic events
	subscribeMemoGroup := v1.Group("/dynamic-events")
	{
		subscribeMemoGroup.POST("", h.DynamicEvents.Events)
	}

	/////////////////
	// PUBLIC API GROUP
	/////////////////

	// assets
	publicGroup := v1.Group("/public")
	{
		publicGroup.GET("/employees", h.Employee.PublicList)
		publicGroup.GET("/clients", h.Client.PublicList)
		publicGroup.GET("/community-nfts/:id", h.CommunityNft.GetNftMetadata)
		publicGroup.GET("/youtube/broadcast", h.Youtube.LatestBroadcast)
	}
}
