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
	pmw := mw.NewPermissionMiddleware(s, repo, cfg)
	amw := mw.NewAuthMiddleware(cfg, s, repo)

	/////////////////
	// Cronjob GROUP
	/////////////////
	cronjob := r.Group("/cronjobs")
	{
		cronjob.POST("/audits", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Audit.Sync)
		cronjob.POST("/birthday", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Discord.BirthdayDailyMessage)
		cronjob.POST("/on-leaves", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Discord.OnLeaveMessage)
		cronjob.POST("/sync-discord-info", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Discord.SyncDiscordInfo)
		cronjob.POST("/sync-monthly-accounting-todo", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Accounting.CreateAccountingTodo)
		cronjob.POST("/sync-project-member-status", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Project.SyncProjectMemberStatus)
		cronjob.POST("/store-vault-transaction", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Vault.StoreVaultTransaction)
		cronjob.POST("/index-engagement-messages", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Engagement.IndexMessages)
		cronjob.POST("/brainery-reports", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Discord.ReportBraineryMetrics)
		cronjob.POST("/delivery-metric-reports", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Discord.DeliveryMetricsReport)
		cronjob.POST("/sync-delivery-metrics", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.DeliveryMetric.Sync)
		cronjob.POST("/sync-conversion-rates", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.ConversionRate.Sync)
		cronjob.POST("/sync-memo", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Discord.SyncMemo)
		cronjob.POST("/sweep-memo", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Discord.SweepMemo)
		cronjob.POST("/notify-weekly-memos", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Discord.NotifyWeeklyMemos)
		cronjob.POST("/notify-top-memo-authors", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Discord.NotifyTopMemoAuthors)
		cronjob.POST("/transcribe-youtube-broadcast", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Youtube.TranscribeBroadcast)
		cronjob.POST("/sweep-ogif-event", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.Discord.SweepOgifEvent)
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
		assetGroup.POST("/upload", amw.WithAuth, h.Asset.Upload)
	}

	lineManagerGroup := v1.Group("/line-managers")
	{
		lineManagerGroup.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesRead), h.Employee.GetLineManagers)
	}

	// auth
	authRoute := v1.Group("/auth")
	{
		authRoute.POST("", h.Auth.Auth)
		authRoute.GET("/me", amw.WithAuth, pmw.WithPerm(model.PermissionAuthRead), h.Auth.Me)
		authRoute.POST("/api-key", amw.WithAuth, pmw.WithPerm(model.PermissionAuthCreate), h.Auth.CreateAPIKey)
	}

	// user profile
	profileGroup := v1.Group("/profile")
	{
		profileGroup.GET("", amw.WithAuth, h.Profile.GetProfile)
		profileGroup.PUT("", amw.WithAuth, h.Profile.UpdateInfo)
		profileGroup.POST("/upload-avatar", amw.WithAuth, h.Profile.UploadAvatar)
		profileGroup.POST("/upload", amw.WithAuth, h.Profile.Upload)
	}

	// employees
	employeeRoute := v1.Group("/employees")
	{
		employeeRoute.POST("", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesCreate), h.Employee.Create)
		employeeRoute.POST("/search", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesRead), h.Employee.List)
		employeeRoute.GET("/:id", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesRead), h.Employee.Details)
		employeeRoute.PUT("/:id/general-info", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdateGeneralInfo)
		employeeRoute.PUT("/:id/personal-info", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdatePersonalInfo)
		employeeRoute.PUT("/:id/skills", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdateSkills)
		employeeRoute.PUT("/:id/employee-status", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdateEmployeeStatus)
		employeeRoute.POST("/:id/upload-avatar", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UploadAvatar)
		employeeRoute.PUT("/:id/roles", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeeRolesEdit), h.Employee.UpdateRole)
		employeeRoute.PUT("/:id/base-salary", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesBaseSalaryEdit), h.Employee.UpdateBaseSalary)
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
		metadataRoute.PUT("/stacks/:id", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataEdit), h.Metadata.UpdateStack)
		metadataRoute.POST("/stacks", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataCreate), h.Metadata.CreateStack)
		metadataRoute.DELETE("/stacks/:id", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataDelete), h.Metadata.DeleteStack)
		metadataRoute.PUT("/positions/:id", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataEdit), h.Metadata.UpdatePosition)
		metadataRoute.POST("/positions", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataCreate), h.Metadata.CreatePosition)
		metadataRoute.DELETE("/positions/:id", amw.WithAuth, pmw.WithPerm(model.PermissionMetadataDelete), h.Metadata.DeletePosition)
	}

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
		projectGroup.DELETE("/:id/members/:memberID", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersDelete), h.Project.DeleteMember)
		projectGroup.DELETE("/:id/slots/:slotID", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersDelete), h.Project.DeleteSlot)
		projectGroup.PUT("/:id/general-info", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UpdateGeneralInfo)
		projectGroup.PUT("/:id/contact-info", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UpdateContactInfo)
		projectGroup.GET("/:id/work-units", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsRead), h.Project.GetWorkUnits)
		projectGroup.POST("/:id/work-units", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsCreate), h.Project.CreateWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsEdit), h.Project.UpdateWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID/archive", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsEdit), h.Project.ArchiveWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID/unarchive", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsEdit), h.Project.UnarchiveWorkUnit)
		projectGroup.GET("/:id/commission-models", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsCommissionModelsRead), h.Project.CommissionModels)
		projectGroup.GET("/icy-distribution/weekly", amw.WithAuth, pmw.WithPerm(model.PermissionIcyDistributionRead), h.Project.IcyWeeklyDistribution)
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

	companyInfoGroup := v1.Group("/company-infos")
	{
		companyInfoGroup.GET("", pmw.WithPerm(model.PermissionCompanyInfoRead), h.CompanyInfo.List)
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

	invitationGroup := v1.Group("/invite")
	{
		invitationGroup.GET("", amw.WithAuth, h.Profile.GetInvitation)
		invitationGroup.PUT("/submit", amw.WithAuth, h.Profile.SubmitOnboardingForm)
	}

	engagementsGroup := v1.Group("/engagements")
	{
		engagementsGroup.POST(
			"/rollup",
			amw.WithAuth,
			pmw.WithPerm(model.PermissionEngagementMetricsWrite),
			h.Engagement.UpsertRollup,
		)
		engagementsGroup.GET(
			"/channels/:channel-id/last-message-id",
			amw.WithAuth,
			pmw.WithPerm(model.PermissionEngagementMetricsRead),
			h.Engagement.GetLastMessageID,
		)
	}

	braineryGroup := v1.Group("/brainery-logs")
	{
		braineryGroup.POST("", amw.WithAuth, pmw.WithPerm(model.PermissionBraineryLogsWrite), h.BraineryLog.Create)
		braineryGroup.GET("/metrics", amw.WithAuth, pmw.WithPerm(model.PermissionBraineryLogsRead), h.BraineryLog.GetMetrics)
		braineryGroup.POST("/sync", amw.WithAuth, pmw.WithPerm(model.PermissionCronjobExecute), h.BraineryLog.Sync)
	}

	memoGroup := v1.Group("/memos")
	{
		memoGroup.POST("", amw.WithAuth, h.MemoLog.Create)
		memoGroup.POST("/sync", amw.WithAuth, h.MemoLog.Sync)
		memoGroup.GET("", amw.WithAuth, h.MemoLog.List)
		memoGroup.GET("/discords", amw.WithAuth, h.MemoLog.ListByDiscordID)
		memoGroup.GET("/prs", amw.WithAuth, h.MemoLog.ListOpenPullRequest)
		memoGroup.GET("/top-authors", amw.WithAuth, h.MemoLog.GetTopAuthors)
	}

	earnGroup := v1.Group("/earns")
	{
		earnGroup.GET("", amw.WithAuth, h.Earn.ListEarn)
	}

	// Delivery metrics
	{
		deliveryGroup := v1.Group("/delivery-metrics")
		deliveryGroup.POST("/report/sync", amw.WithAuth, pmw.WithPerm(model.PermissionDeliveryMetricsSync), h.DeliveryMetric.Sync)

		deliveryGroup.GET("/report/weekly", amw.WithAuth, pmw.WithPerm(model.PermissionDeliveryMetricsRead), h.DeliveryMetric.GetWeeklyReport)
		deliveryGroup.GET("/report/monthly", amw.WithAuth, pmw.WithPerm(model.PermissionDeliveryMetricsRead), h.DeliveryMetric.GetMonthlyReport)
		deliveryGroup.GET("/leader-board/weekly", amw.WithAuth, pmw.WithPerm(model.PermissionDeliveryMetricsLeaderBoardRead), h.DeliveryMetric.GetWeeklyLeaderBoard)
		deliveryGroup.GET("/leader-board/monthly", amw.WithAuth, pmw.WithPerm(model.PermissionDeliveryMetricsLeaderBoardRead), h.DeliveryMetric.GetMonthlyLeaderBoard)

		// API for fortress-discord
		deliveryGroup.GET("/report/weekly/discord-msg", amw.WithAuth, pmw.WithPerm(model.PermissionDeliveryMetricsRead), h.DeliveryMetric.GetWeeklyReportDiscordMsg)
		deliveryGroup.GET("/report/monthly/discord-msg", amw.WithAuth, pmw.WithPerm(model.PermissionDeliveryMetricsRead), h.DeliveryMetric.GetMonthlyReportDiscordMsg)
	}

	discordGroup := v1.Group("/discords")
	{
		discordGroup.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Employee.ListByDiscordRequest)
		discordGroup.GET("/mma-scores", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Employee.ListWithMMAScore)
		discordGroup.POST("/advance-salary", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Employee.SalaryAdvance)
		discordGroup.POST("/check-advance-salary", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Employee.CheckSalaryAdvance)

		discordGroup.GET("/salary-advance-report", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Employee.SalaryAdvanceReport)
		discordGroup.GET("/:discord_id/earns/transactions", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Employee.GetEmployeeEarnTransactions)
		discordGroup.GET("/:discord_id/earns/total", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Employee.GetEmployeeTotalEarn)
		discordGroup.GET("/earns/total", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Employee.GetTotalEarn)
		discordGroup.POST("/office-checkin", amw.WithAuth, pmw.WithPerm(model.PermissionTransferCheckinIcy), h.Employee.OfficeCheckIn)

		discordGroup.GET("/icy-accounting", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Icy.Accounting)

		scheduledEventGroup := discordGroup.Group("/scheduled-events")
		{
			scheduledEventGroup.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Discord.ListScheduledEvent)
			scheduledEventGroup.POST("", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordCreate), h.Discord.CreateScheduledEvent)
			scheduledEventGroup.PUT("/:id/speakers", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordEdit), h.Discord.SetScheduledEventSpeakers)
		}

		discordGroup.GET("/research-topics", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Discord.ListDiscordResearchTopics)
	}

	conversionRateGroup := v1.Group("/conversion-rates")
	{
		conversionRateGroup.GET("", amw.WithAuth, h.ConversionRate.List)
	}

	newsGroup := v1.Group("/news")
	{
		newsGroup.GET("", amw.WithAuth, h.News.Fetch)
	}

	ogifGroup := v1.Group("/ogif")
	{
		ogifGroup.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Discord.UserOgifStats)
		ogifGroup.GET("/leaderboard", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesDiscordRead), h.Discord.OgifLeaderboard)
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
