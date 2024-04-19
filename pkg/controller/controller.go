package controller

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller/auth"
	"github.com/dwarvesf/fortress-api/pkg/controller/brainerylogs"
	"github.com/dwarvesf/fortress-api/pkg/controller/client"
	"github.com/dwarvesf/fortress-api/pkg/controller/conversionrate"
	"github.com/dwarvesf/fortress-api/pkg/controller/deliverymetrics"
	"github.com/dwarvesf/fortress-api/pkg/controller/discord"
	"github.com/dwarvesf/fortress-api/pkg/controller/employee"
	"github.com/dwarvesf/fortress-api/pkg/controller/icy"
	"github.com/dwarvesf/fortress-api/pkg/controller/invoice"
	"github.com/dwarvesf/fortress-api/pkg/controller/memologs"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

type Controller struct {
	Auth           auth.IController
	BraineryLog    brainerylogs.IController
	Client         client.IController
	ConversionRate conversionrate.IController
	DeliveryMetric deliverymetrics.IController
	Employee       employee.IController
	Invoice        invoice.IController
	Discord        discord.IController
	Icy            icy.IController
	MemoLog        memologs.IController
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, worker *worker.Worker, logger logger.Logger, cfg *config.Config) *Controller {
	return &Controller{
		Auth:           auth.New(store, repo, service, logger, cfg),
		BraineryLog:    brainerylogs.New(store, repo, service, logger, cfg),
		Client:         client.New(store, repo, service, logger, cfg),
		ConversionRate: conversionrate.New(store, repo, service, logger, cfg),
		DeliveryMetric: deliverymetrics.New(store, repo, service, logger, cfg),
		Employee:       employee.New(store, repo, service, logger, cfg),
		Invoice:        invoice.New(store, repo, service, worker, logger, cfg),
		Discord:        discord.New(store, repo, service, logger, cfg),
		Icy:            icy.New(service, logger, cfg),
		MemoLog:        memologs.New(store, repo, service, logger, cfg),
	}
}
