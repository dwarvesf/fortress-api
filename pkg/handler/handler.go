package handler

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/accounting"
	"github.com/dwarvesf/fortress-api/pkg/handler/asset"
	"github.com/dwarvesf/fortress-api/pkg/handler/audit"
	"github.com/dwarvesf/fortress-api/pkg/handler/auth"
	"github.com/dwarvesf/fortress-api/pkg/handler/bankaccount"
	"github.com/dwarvesf/fortress-api/pkg/handler/brainerylogs"
	"github.com/dwarvesf/fortress-api/pkg/handler/client"
	"github.com/dwarvesf/fortress-api/pkg/handler/communitynft"
	"github.com/dwarvesf/fortress-api/pkg/handler/companyinfo"
	"github.com/dwarvesf/fortress-api/pkg/handler/conversionrate"
	"github.com/dwarvesf/fortress-api/pkg/handler/dashboard"
	"github.com/dwarvesf/fortress-api/pkg/handler/dashboard/util"
	"github.com/dwarvesf/fortress-api/pkg/handler/deliverymetric"
	"github.com/dwarvesf/fortress-api/pkg/handler/discord"
	"github.com/dwarvesf/fortress-api/pkg/handler/earn"
	"github.com/dwarvesf/fortress-api/pkg/handler/employee"
	"github.com/dwarvesf/fortress-api/pkg/handler/engagement"
	"github.com/dwarvesf/fortress-api/pkg/handler/feedback"
	"github.com/dwarvesf/fortress-api/pkg/handler/healthz"
	"github.com/dwarvesf/fortress-api/pkg/handler/icy"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice"
	"github.com/dwarvesf/fortress-api/pkg/handler/memologs"
	"github.com/dwarvesf/fortress-api/pkg/handler/metadata"
	"github.com/dwarvesf/fortress-api/pkg/handler/news"
	"github.com/dwarvesf/fortress-api/pkg/handler/notion"
	"github.com/dwarvesf/fortress-api/pkg/handler/payroll"
	"github.com/dwarvesf/fortress-api/pkg/handler/profile"
	"github.com/dwarvesf/fortress-api/pkg/handler/project"
	"github.com/dwarvesf/fortress-api/pkg/handler/survey"
	"github.com/dwarvesf/fortress-api/pkg/handler/valuation"
	"github.com/dwarvesf/fortress-api/pkg/handler/vault"
	"github.com/dwarvesf/fortress-api/pkg/handler/webhook"
	yt "github.com/dwarvesf/fortress-api/pkg/handler/youtube"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/worker"
	"github.com/dwarvesf/fortress-api/pkg/handler/dynamicevents"
)

type Handler struct {
	Accounting     accounting.IHandler
	Asset          asset.IHandler
	Audit          audit.IHandler
	Auth           auth.IHandler
	BankAccount    bankaccount.IHandler
	BraineryLog    brainerylogs.IHandler
	Client         client.IHandler
	CompanyInfo    companyinfo.IHandler
	ConversionRate conversionrate.IHandler
	Dashboard      dashboard.IHandler
	DeliveryMetric deliverymetric.IHandler
	Discord        discord.IHandler
	Employee       employee.IHandler
	Engagement     engagement.IHandler
	Feedback       feedback.IHandler
	Healthcheck    healthz.IHandler
	Invoice        invoice.IHandler
	MemoLog        memologs.IHandler
	Metadata       metadata.IHandler
	Notion         notion.IHandler
	Payroll        payroll.IHandler
	Profile        profile.IHandler
	Project        project.IHandler
	Survey         survey.IHandler
	Valuation      valuation.IHandler
	Webhook        webhook.IHandler
	Vault          vault.IHandler
	Icy            icy.IHandler
	CommunityNft   communitynft.IHandler
	Earn           earn.IHandler
	News           news.IHandler
	Youtube        yt.IHandler
	DynamicEvents  dynamicevents.IHandler
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, ctrl *controller.Controller, worker *worker.Worker, logger logger.Logger, cfg *config.Config) *Handler {
	return &Handler{
		Accounting:     accounting.New(store, repo, service, logger, cfg),
		Asset:          asset.New(store, repo, service, logger, cfg),
		Audit:          audit.New(store, repo, service, logger, cfg),
		Auth:           auth.New(ctrl, logger, cfg),
		BankAccount:    bankaccount.New(store, repo, service, logger, cfg),
		BraineryLog:    brainerylogs.New(ctrl, store, repo, service, logger, cfg),
		Client:         client.New(ctrl, store, repo, service, logger, cfg),
		CompanyInfo:    companyinfo.New(ctrl, store, repo, service, logger, cfg),
		ConversionRate: conversionrate.New(ctrl, store, repo, service, logger, cfg),
		Dashboard:      dashboard.New(store, repo, service, logger, cfg, util.New()),
		DeliveryMetric: deliverymetric.New(ctrl, store, repo, service, logger, cfg),
		Discord:        discord.New(ctrl, store, repo, service, logger, cfg),
		Employee:       employee.New(ctrl, store, repo, service, logger, cfg),
		Engagement:     engagement.New(ctrl, store, repo, service, logger, cfg),
		Feedback:       feedback.New(store, repo, service, logger, cfg),
		Healthcheck:    healthz.New(),
		Invoice:        invoice.New(ctrl, store, repo, service, worker, logger, cfg),
		MemoLog:        memologs.New(ctrl, store, repo, service, logger, cfg),
		Metadata:       metadata.New(store, repo, service, logger, cfg),
		Notion:         notion.New(store, repo, service, logger, cfg),
		Payroll:        payroll.New(ctrl, store, repo, service, worker, logger, cfg),
		Profile:        profile.New(ctrl, store, repo, service, logger, cfg),
		Project:        project.New(ctrl, store, repo, service, logger, cfg),
		Survey:         survey.New(store, repo, service, logger, cfg),
		Valuation:      valuation.New(store, repo, service, logger, cfg),
		Webhook:        webhook.New(ctrl, store, repo, service, logger, cfg, worker),
		Vault:          vault.New(store, repo, service, logger, cfg),
		Icy:            icy.New(ctrl, logger),
		CommunityNft:   communitynft.New(ctrl, store, repo, service, logger, cfg),
		Earn:           earn.New(ctrl, store, repo, service, logger, cfg),
		News:           news.New(store, repo, ctrl, logger, cfg),
		Youtube:        yt.New(ctrl, store, repo, service, logger, cfg),
		DynamicEvents:  dynamicevents.New(store, repo, ctrl, logger, cfg, service),
	}
}
