package withdrawal

import (
	mochisdk "github.com/consolelabs/mochi-go-sdk/mochi"
	mochicfg "github.com/consolelabs/mochi-go-sdk/mochi/config"
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
)

type IController interface {
	CheckWithdrawalCondition(in CheckWithdrawInput) (*model.WithdrawalCondition, error)
	RequestTransfer(in WithdrawInput) (string, error)
	ConfirmTransaction(in ConfirmTransactionInput) error
}

type controller struct {
	service *service.Service
	// mochiAppClient is mochi withdraw money app client
	mochiAppClient mochisdk.APIClient
	logger         logger.Logger
	config         *config.Config
}

func New(service *service.Service, logger logger.Logger, cfg *config.Config) IController {
	return &controller{
		service: service,
		logger:  logger,
		config:  cfg,
		mochiAppClient: mochisdk.NewClient(
			&mochicfg.Config{
				MochiPay: mochicfg.MochiPay{
					BaseURL:         cfg.MochiWithdraw.BaseURL,
					ApplicationID:   cfg.MochiWithdraw.ApplicationID,
					ApplicationName: cfg.MochiWithdraw.ApplicationName,
					APIKey:          cfg.MochiWithdraw.APIKey,
					IsPreview:       false,
				},
				MochiProfile: mochicfg.MochiProfile{},
			},
		),
	}
}
