package withdrawal

import (
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

type CheckWithdrawInput struct {
	DiscordID string `json:"discordID"`
}

const icyUsdRate = 1.5

// CheckWithdrawalCondition means check condition to withdraw.
func (c *controller) CheckWithdrawalCondition(in CheckWithdrawInput) (*model.WithdrawalCondition, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "withdrawal",
		"method":     "Withdraw",
	})

	profile, err := c.mochiAppClient.GetByDiscordID(in.DiscordID)
	if err != nil {
		l.Error(err, "failed to get mochi profile")
		return nil, err
	}

	userICYBalance, err := c.mochiAppClient.GetUserBalances(profile.ID)
	if err != nil {
		l.Error(err, "failed to get user balance")
		return nil, err
	}

	currentICYAmount := utils.ConvertFromString(userICYBalance[0].Amount, int64(userICYBalance[0].Token.Decimal))
	currentICYAmount.SetPrec(5)

	// Get current ICY/USD rate
	icyAmount, _ := currentICYAmount.Float64()
	amountUSD := icyAmount * icyUsdRate
	convertedICYAmount, rate, err := c.service.Wise.Convert(amountUSD, "USD", "VND")
	if err != nil {
		return nil, err
	}

	return &model.WithdrawalCondition{
		ICYVNDRate: rate * icyUsdRate,
		ICYAmount:  icyAmount,
		VNDAmount:  convertedICYAmount,
	}, nil
}
