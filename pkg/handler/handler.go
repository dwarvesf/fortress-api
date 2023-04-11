package handler

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/accounting"
	"github.com/dwarvesf/fortress-api/pkg/handler/asset"
	"github.com/dwarvesf/fortress-api/pkg/handler/audience"
	"github.com/dwarvesf/fortress-api/pkg/handler/audit"
	"github.com/dwarvesf/fortress-api/pkg/handler/auth"
	"github.com/dwarvesf/fortress-api/pkg/handler/bankaccount"
	"github.com/dwarvesf/fortress-api/pkg/handler/birthday"
	"github.com/dwarvesf/fortress-api/pkg/handler/changelog"
	"github.com/dwarvesf/fortress-api/pkg/handler/client"
	"github.com/dwarvesf/fortress-api/pkg/handler/dashboard"
	"github.com/dwarvesf/fortress-api/pkg/handler/dashboard/util"
	"github.com/dwarvesf/fortress-api/pkg/handler/dfupdate"
	"github.com/dwarvesf/fortress-api/pkg/handler/digest"
	"github.com/dwarvesf/fortress-api/pkg/handler/discord"
	"github.com/dwarvesf/fortress-api/pkg/handler/earn"
	"github.com/dwarvesf/fortress-api/pkg/handler/employee"
	"github.com/dwarvesf/fortress-api/pkg/handler/event"
	"github.com/dwarvesf/fortress-api/pkg/handler/feedback"
	"github.com/dwarvesf/fortress-api/pkg/handler/healthz"
	"github.com/dwarvesf/fortress-api/pkg/handler/hiring"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice"
	"github.com/dwarvesf/fortress-api/pkg/handler/issue"
	"github.com/dwarvesf/fortress-api/pkg/handler/memo"
	"github.com/dwarvesf/fortress-api/pkg/handler/metadata"
	"github.com/dwarvesf/fortress-api/pkg/handler/payroll"
	"github.com/dwarvesf/fortress-api/pkg/handler/profile"
	"github.com/dwarvesf/fortress-api/pkg/handler/project"
	"github.com/dwarvesf/fortress-api/pkg/handler/staffingdemand"
	"github.com/dwarvesf/fortress-api/pkg/handler/survey"
	"github.com/dwarvesf/fortress-api/pkg/handler/techradar"
	"github.com/dwarvesf/fortress-api/pkg/handler/update"
	"github.com/dwarvesf/fortress-api/pkg/handler/valuation"
	"github.com/dwarvesf/fortress-api/pkg/handler/webhook"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

type Handler struct {
	Healthcheck    healthz.IHandler
	Employee       employee.IHandler
	Metadata       metadata.IHandler
	Auth           auth.IHandler
	Project        project.IHandler
	Profile        profile.IHandler
	Feedback       feedback.IHandler
	Survey         survey.IHandler
	Dashboard      dashboard.IHandler
	Valuation      valuation.IHandler
	Earn           earn.IHandler
	TechRadar      techradar.IHandler
	Issue          issue.IHandler
	Audience       audience.IHandler
	Event          event.IHandler
	Hiring         hiring.IHandler
	StaffingDemand staffingdemand.IHandler
	Audit          audit.IHandler
	Digest         digest.IHandler
	Update         update.IHandler
	Memo           memo.IHandler
	BankAccount    bankaccount.IHandler
	Birthday       birthday.IHandler
	Invoice        invoice.IHandler
	Webhook        webhook.IHandler
	Discord        discord.IHandler
	Client         client.IHandler
	Asset          asset.IHandler
	Changelog      changelog.IHandler
	DFUpdate       dfupdate.IHandler
	Payroll        payroll.IHandler
	Accounting     accounting.IHandler
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, ctrl *controller.Controller, worker *worker.Worker, logger logger.Logger, cfg *config.Config) *Handler {
	return &Handler{
		Healthcheck:    healthz.New(),
		Employee:       employee.New(ctrl, store, repo, service, logger, cfg),
		Metadata:       metadata.New(store, repo, service, logger, cfg),
		Auth:           auth.New(ctrl, logger, cfg),
		Project:        project.New(store, repo, service, logger, cfg),
		Profile:        profile.New(store, repo, service, logger, cfg),
		Feedback:       feedback.New(store, repo, service, logger, cfg),
		Survey:         survey.New(store, repo, service, logger, cfg),
		Dashboard:      dashboard.New(store, repo, service, logger, cfg, util.New()),
		Valuation:      valuation.New(store, repo, service, logger, cfg),
		Earn:           earn.New(store, repo, service, logger, cfg),
		TechRadar:      techradar.New(store, repo, service, logger, cfg),
		Issue:          issue.New(store, repo, service, logger, cfg),
		Audience:       audience.New(store, repo, service, logger, cfg),
		Event:          event.New(store, repo, service, logger, cfg),
		Hiring:         hiring.New(store, repo, service, logger, cfg),
		StaffingDemand: staffingdemand.New(store, repo, service, logger, cfg),
		Audit:          audit.New(store, repo, service, logger, cfg),
		Digest:         digest.New(store, repo, service, logger, cfg),
		Update:         update.New(store, repo, service, logger, cfg),
		Memo:           memo.New(store, repo, service, logger, cfg),
		BankAccount:    bankaccount.New(store, repo, service, logger, cfg),
		Birthday:       birthday.New(store, repo, service, logger, cfg),
		Invoice:        invoice.New(ctrl, store, repo, service, worker, logger, cfg),
		Webhook:        webhook.New(store, repo, service, logger, cfg),
		Discord:        discord.New(store, repo, service, logger, cfg),
		Client:         client.New(ctrl, store, repo, service, logger, cfg),
		Asset:          asset.New(store, repo, service, logger, cfg),
		Changelog:      changelog.New(store, repo, service, logger, cfg),
		DFUpdate:       dfupdate.New(store, repo, service, logger, cfg),
		Payroll:        payroll.New(store, repo, service, worker, logger, cfg),
		Accounting:     accounting.New(store, repo, service, logger, cfg),
	}
}
