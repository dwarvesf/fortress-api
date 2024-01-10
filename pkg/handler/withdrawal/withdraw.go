package withdrawal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/controller"
	withdrawalCtrl "github.com/dwarvesf/fortress-api/pkg/controller/withdrawal"
	"github.com/dwarvesf/fortress-api/pkg/handler/withdrawal/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	controller *controller.Controller
	logger     logger.Logger
}

func New(
	controller *controller.Controller,
	logger logger.Logger,
) IHandler {
	return &handler{
		controller: controller,
		logger:     logger,
	}
}

// CheckWithdrawCondition means check condition to withdraw money from ICY
// @Summary Check withdraw condition by discord id
// @Description Check withdraw condition by discord id
// @id checkWithdrawCondition
// @Tags Withdraw
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param discordID query string false "DiscordID"
// @Success 200 {object} CheckWithdrawConditionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /employees/check-advance-salary [post]
func (h *handler) CheckWithdrawCondition(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "icy",
		"method":  "Balance",
	})

	query := request.CheckWithdrawConditionRequest{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	rs, err := h.controller.Withdrawal.CheckWithdrawalCondition(withdrawalCtrl.CheckWithdrawInput{
		DiscordID: query.DiscordID,
	})
	if err != nil {
		l.Error(err, "failed to check withdraw condition")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, "failed to check with withdraw condition"))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToCheckWithdrawCondition(rs), nil, nil, nil, ""))
}

//
//func (h *handler) WithdrawMoney(c *gin.Context) {
//	l := h.logger.Fields(logger.Fields{
//		"handler": "icy",
//		"method":  "Withdraw",
//	})
//
//	var req request.WithdrawMoneyRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		l.Error(err, "failed to bind json")
//		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "failed to bind json"))
//		return
//	}
//}
