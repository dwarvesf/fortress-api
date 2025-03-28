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
		}).Error(err, "request is existed")
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
		SwapRequestStatus: model.SwapRequestStatusPending,
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

	// implement swap icy to btc
	swapTx, errSwap := h.controller.Swap.Swap(&transferRequest)
	if errSwap != nil {
		revertStatus := model.RevertRequestStatusSuccess
		txDeposit, errRevertIcy := h.controller.Swap.RevertIcyToUser(&transferRequest)
		if errRevertIcy != nil {
			revertStatus = model.RevertRequestStatusFailed
			icySwapBtcRequest.RevertError = errRevertIcy.Error()
		}

		// update icy swap btc request failed
		icySwapBtcRequest.SwapRequestStatus = model.SwapRequestStatusFailed
		icySwapBtcRequest.RevertStatus = revertStatus
		icySwapBtcRequest.TxDeposit = txDeposit
		icySwapBtcRequest.SwapRequestError = errSwap.Error()
		_, err = h.store.IcySwapBtcRequest.Update(h.repo.DB(), icySwapBtcRequest)
		if err != nil {
			h.logger.Fields(logger.Fields{
				"data":  req,
				"event": "Transfer Request",
			}).Error(err, "can't update failed swap request ")
			return
		}

		return
	}

	// update icy swap btc request success
	icySwapBtcRequest.SwapRequestStatus = model.SwapRequestStatusSuccess
	icySwapBtcRequest.TxSwap = swapTx
	_, err = h.store.IcySwapBtcRequest.Update(h.repo.DB(), icySwapBtcRequest)
	if err != nil {
		h.logger.Fields(logger.Fields{
			"data":  req,
			"event": "Transfer Request",
		}).Error(err, "can't update success swap request")
		return
	}
}
