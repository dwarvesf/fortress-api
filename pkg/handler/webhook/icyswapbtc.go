package webhook

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

const (
	RequestStatus = "success"
)

func (h *handler) TransferRequest(c *gin.Context) {
	var req TransactionRequestEvent

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// Return success response immediately
	c.JSON(http.StatusAccepted, view.CreateResponse(gin.H{
		"request_code": req.RequestCode,
		"status":       "processing",
	}, nil, nil, nil, "Request accepted and processing"))

	// Continue processing in a goroutine
	go func() {
		h.processTransferRequest(req)
	}()
}

// New function to handle the asynchronous processing
func (h *handler) processTransferRequest(req TransactionRequestEvent) {
	// check approve transfer icy request
	if req.Status != RequestStatus {
		h.logger.Fields(logger.Fields{
			"data":  req,
			"event": "Transfer Request",
		}).Error(errors.New("invalid request"), "transfer request is invalid")
		return
	}

	requestIsExist, err := h.store.IcySwapBtcRequest.IsExist(h.repo.DB(), req.RequestCode)
	if err != nil {
		h.logger.Fields(logger.Fields{
			"data":  req,
			"event": "Transfer Request",
		}).Error(err, "can't check request existed")
		return
	}

	if requestIsExist {
		h.logger.Fields(logger.Fields{
			"data":  req,
			"event": "Transfer Request",
		}).Error(errors.New("request already exists"), "request is existed")
		return
	}

	// init icy swap btc request record
	icySwapBtcRequest, err := h.store.IcySwapBtcRequest.Create(h.repo.DB(), &model.IcySwapBtcRequest{
		ProfileID:         req.ProfileID,
		RequestCode:       req.RequestCode,
		Amount:            req.Amount,
		TokenID:           req.TokenID,
		Timestamp:         req.Timestamp,
		TokenName:         req.TokenName,
		SwapRequestStatus: model.IcySwapBtcStatusPending,
		BtcAddress:        req.Description,
		TxStatus:          req.Status,
		TxID:              req.TxID,
	})
	if err != nil {
		h.logger.Fields(logger.Fields{
			"data":  req,
			"event": "Transfer Request",
		}).Error(err, "can't store request")
		return
	}

	transferRequest := model.TransferRequestResponse{
		ProfileID:   req.ProfileID,
		Description: req.Description,
		RequestCode: req.RequestCode,
		Amount:      req.Amount,
		TokenID:     req.TokenID,
		Timestamp:   req.Timestamp,
		TokenName:   req.TokenName,
		Status:      req.Status,
		TxID:        req.TxID,
	}

	// implement withdraw from vault
	err = h.controller.Swap.WithdrawFromVault(&transferRequest)
	if err != nil {
		h.logger.Fields(logger.Fields{
			"data":  req,
			"event": "Transfer Request",
		}).Error(err, "can't withdraw from vault")

		icySwapBtcRequest.WithdrawStatus = model.IcySwapBtcStatusFailed
		icySwapBtcRequest.WithdrawError = err.Error()
		icySwapBtcRequest.SwapRequestStatus = model.IcySwapBtcStatusFailed
		icySwapBtcRequest.SwapRequestError = err.Error()

		revertStatus := model.IcySwapBtcStatusSuccess
		errT := h.controller.Swap.TransferFromVaultToUser(&transferRequest)
		if errT != nil {
			revertStatus = model.IcySwapBtcStatusFailed
			h.logger.Fields(logger.Fields{
				"data":  req,
				"event": "Transfer Request",
			}).Error(errT, "can't transfer from vault to user")
		}

		icySwapBtcRequest.RevertStatus = revertStatus
		_, errStore := h.store.IcySwapBtcRequest.Update(h.repo.DB(), icySwapBtcRequest)
		if errStore != nil {
			h.logger.Fields(logger.Fields{
				"data":  req,
				"event": "Transfer Request",
			}).Error(errStore, "can't update failed swap request ")
		}
		return
	}

	icySwapBtcRequest.WithdrawStatus = model.IcySwapBtcStatusSuccess

	// implement swap icy to btc
	swapTx, errSwap := h.controller.Swap.Swap(&transferRequest)
	if errSwap != nil {
		icySwapBtcRequest.SwapRequestError = errSwap.Error()
		icySwapBtcRequest.SwapRequestStatus = model.IcySwapBtcStatusFailed
		revertStatus := model.IcySwapBtcStatusSuccess

		txDeposit, errDeposit := h.controller.Swap.DepositToVault(&transferRequest)
		if errDeposit != nil {
			icySwapBtcRequest.RevertError = errDeposit.Error()
			icySwapBtcRequest.RevertStatus = model.IcySwapBtcStatusFailed
			icySwapBtcRequest.TxDeposit = txDeposit

			_, errStore := h.store.IcySwapBtcRequest.Update(h.repo.DB(), icySwapBtcRequest)
			if errStore != nil {
				h.logger.Fields(logger.Fields{
					"data":  req,
					"event": "Transfer Request",
				}).Error(errStore, "can't update failed swap request ")
			}
			return
		}

		// transfer icy from vault to user
		errT := h.controller.Swap.TransferFromVaultToUser(&transferRequest)
		if errT != nil {
			h.logger.Fields(logger.Fields{
				"data":  req,
				"event": "Transfer Request",
			}).Error(errT, "can't transfer from vault to user")

			revertStatus = model.IcySwapBtcStatusFailed
			icySwapBtcRequest.RevertError = errT.Error()
		}

		icySwapBtcRequest.RevertStatus = revertStatus
		icySwapBtcRequest.TxDeposit = txDeposit

		_, errStore := h.store.IcySwapBtcRequest.Update(h.repo.DB(), icySwapBtcRequest)
		if errStore != nil {
			h.logger.Fields(logger.Fields{
				"data":  req,
				"event": "Transfer Request",
			}).Error(errStore, "can't update failed swap request ")
		}
		return
	}

	// update icy swap btc request success
	icySwapBtcRequest.SwapRequestStatus = model.IcySwapBtcStatusSuccess
	icySwapBtcRequest.TxSwap = swapTx
	_, err = h.store.IcySwapBtcRequest.Update(h.repo.DB(), icySwapBtcRequest)
	if err != nil {
		h.logger.Fields(logger.Fields{
			"data":  req,
			"event": "Transfer Request",
		}).Error(err, "can't update success swap request")
	}
}
