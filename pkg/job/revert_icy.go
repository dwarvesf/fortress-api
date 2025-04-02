package job

import (
	"fmt"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/icyswapbtc"
)

type revertIcy struct {
	controller *controller.Controller
	service    *service.Service
	store      *store.Store
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
}

func NewRevertIcy(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) Job {
	return &revertIcy{
		controller: controller,
		service:    service,
		store:      store,
		logger:     logger,
		config:     cfg,
		repo:       repo,
	}
}

func (u *revertIcy) Run() error {
	// Create a query to find all requests with RevertStatus == "failed"
	query := &icyswapbtc.Query{
		RevertStatus: model.IcySwapBtcStatusFailed,
	}

	// Get all requests with RevertStatus == "failed"
	requests, err := u.store.IcySwapBtcRequest.All(u.repo.DB(), query)
	if err != nil {
		u.logger.Errorf(err, "[updateRevertIcy] failed to get requests")
		return err
	}

	for i := range requests {
		// Use pointer to the array element to ensure updates affect the original object
		request := &requests[i]

		transferRequest := &model.TransferRequestResponse{
			ProfileID:   request.ProfileID,
			RequestCode: request.RequestCode,
			Status:      request.TxStatus,
			TxID:        request.TxID,
			Description: request.BtcAddress,
			Amount:      request.Amount,
			TokenID:     request.TokenID,
			TokenName:   request.TokenName,
		}

		// Process based on whether TxDeposit exists
		if err := u.processRevertRequest(request, transferRequest); err != nil {
			u.logger.Errorf(err, "[revertIcy] failed to process request %s", request.ID)
			continue
		}
	}

	return nil
}

func (u *revertIcy) processRevertRequest(request *model.IcySwapBtcRequest, transferRequest *model.TransferRequestResponse) error {
	if request.WithdrawStatus == model.IcySwapBtcStatusFailed {
		return u.processTransferFromVaultToUser(request, transferRequest)
	}

	// Empty TxDeposit case
	if request.TxDeposit == "" {
		return u.processDepositToVault(request, transferRequest)
	}

	// Non-empty TxDeposit case
	return u.processTransferFromVaultToUser(request, transferRequest)
}

func (u *revertIcy) processDepositToVault(request *model.IcySwapBtcRequest, transferRequest *model.TransferRequestResponse) error {
	u.logger.Infof("[processDepositToVault] starting")

	// deposit ICY to vault
	txDeposit, err := u.controller.Swap.DepositToVault(transferRequest)
	if err != nil {
		u.logger.Errorf(err, "[processDepositToVault] failed to deposit to vault")
		return err
	}

	// Update the request
	request.TxDeposit = txDeposit
	// Update in database
	_, err = u.store.IcySwapBtcRequest.Update(u.repo.DB(), request)
	if err != nil {
		return fmt.Errorf("[processDepositToVault] failed to update request: %w", err)
	}

	return nil
}

func (u *revertIcy) processTransferFromVaultToUser(request *model.IcySwapBtcRequest, transferRequest *model.TransferRequestResponse) error {
	u.logger.Infof("[processTransferFromVaultToUser] starting")
	// Transfer from vault to user
	err := u.controller.Swap.TransferFromVaultToUser(transferRequest)
	if err != nil {
		u.logger.Errorf(err, "[processTransferFromVaultToUser] failed to transfer from vault to user")
		return err
	}

	// Update request status
	request.RevertStatus = model.IcySwapBtcStatusSuccess

	// Update in database
	_, err = u.store.IcySwapBtcRequest.Update(u.repo.DB(), request)
	if err != nil {
		return fmt.Errorf("[processTransferFromVaultToUser] failed to update request: %w", err)
	}

	return nil
}
