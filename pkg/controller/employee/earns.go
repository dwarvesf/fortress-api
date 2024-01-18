package employee

import (
	"math"
	"math/big"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/mochipay"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

type GetEmployeeEarnTransactionsInput struct {
	model.Pagination
}

func (r *controller) GetEmployeeEarnTransactions(discordID string, input GetEmployeeEarnTransactionsInput) (model.EmployeeEarnTransactions, int64, error) {
	profile, err := r.service.MochiProfile.GetProfileByDiscordID(discordID)
	if err != nil {
		return nil, 0, err
	}

	isSender := false

	txns, err := r.service.MochiPay.GetListTransactions(mochipay.ListTransactionsRequest{
		Type:         mochipay.TransactionTypeReceive,
		ActionList:   []mochipay.TransactionAction{mochipay.TransactionActionVaultTransfer},
		Status:       mochipay.TransactionStatusSuccess,
		TokenAddress: mochipay.ICYAddress,
		ChainIDs:     []string{mochipay.POLYGONChainID},
		ProfileID:    profile.ID,
		Page:         input.Page,
		Size:         input.Size,
		IsSender:     &isSender,
		SortBy:       "created_at-",
	})
	return txns.Data, txns.Pagination.Total, err
}

func (r *controller) GetEmployeeTotalEarn(discordID string) (string, float64, error) {
	profile, err := r.service.MochiProfile.GetProfileByDiscordID(discordID)
	if err != nil {
		return "", 0, err
	}

	isSender := false

	txns, err := r.service.MochiPay.GetListTransactions(mochipay.ListTransactionsRequest{
		Type:         mochipay.TransactionTypeReceive,
		ActionList:   []mochipay.TransactionAction{mochipay.TransactionActionVaultTransfer},
		Status:       mochipay.TransactionStatusSuccess,
		TokenAddress: mochipay.ICYAddress,
		ChainIDs:     []string{mochipay.POLYGONChainID},
		ProfileID:    profile.ID,
		Size:         math.MaxInt64,
		IsSender:     &isSender,
		SortBy:       "created_at-",
	})
	if err != nil {
		return "", 0, err
	}

	earnsICY := big.NewFloat(0)
	earnsUSD := float64(0)
	for _, txn := range txns.Data {
		if txn.Amount != "" && txn.Token != nil {
			earnsICY.Add(earnsICY, utils.ConvertFromString(txn.Amount, txn.Token.Decimal))
			earnsUSD += txn.UsdAmount
		}
	}

	return earnsICY.String(), earnsUSD, nil
}
