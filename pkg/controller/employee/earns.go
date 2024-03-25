package employee

import (
	"math"
	"math/big"
	"time"

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
		ChainIDs:     []string{mochipay.BASEChainID},
		ProfileID:    profile.ID,
		Page:         input.Page,
		Size:         input.Size,
		IsSender:     &isSender,
		SortBy:       "created_at-",
	})
	return txns.Data, txns.Pagination.Total, err
}

func (r *controller) GetEmployeeTotalEarn(discordID string) (string, string, error) {
	profile, err := r.service.MochiProfile.GetProfileByDiscordID(discordID)
	if err != nil {
		return "", "", err
	}

	isSender := false

	txns, err := r.service.MochiPay.GetListTransactions(mochipay.ListTransactionsRequest{
		Type:         mochipay.TransactionTypeReceive,
		ActionList:   []mochipay.TransactionAction{mochipay.TransactionActionVaultTransfer},
		Status:       mochipay.TransactionStatusSuccess,
		TokenAddress: mochipay.ICYAddress,
		ChainIDs:     []string{mochipay.BASEChainID},
		ProfileID:    profile.ID,
		Size:         math.MaxInt64,
		IsSender:     &isSender,
		SortBy:       "created_at-",
	})
	if err != nil {
		return "", "", err
	}

	earnsICY := big.NewFloat(0)
	earnsUSD := big.NewFloat(0)
	for _, txn := range txns.Data {
		if txn.Amount != "" && txn.Token != nil {
			earnsICY.Add(earnsICY, utils.ConvertFromString(txn.Amount, txn.Token.Decimal))
			earnsUSD.Add(earnsUSD, big.NewFloat(txn.UsdAmount))
		}
	}

	return earnsICY.String(), earnsUSD.String(), nil
}

func (r *controller) GetTotalEarn(from, to time.Time) (string, string, error) {
	// Step 1: Call to mochi-api to get vaults by df guild id -> vault ids
	vaults, err := r.service.Mochi.GetListVaults(false)
	if err != nil {
		return "", "", err
	}

	dfVaults := make(map[int64]bool, 0)
	for _, vault := range vaults {
		if vault.GuildID == r.config.Discord.IDs.DwarvesGuild {
			dfVaults[vault.ID] = true
		}
	}

	isSender := true

	// Step 2: Call to mochi-payment to get list txns
	// TODO: add filter for API get txn to filter by time
	txns, err := r.service.MochiPay.GetListTransactions(mochipay.ListTransactionsRequest{
		Type:         mochipay.TransactionTypeReceive,
		ActionList:   []mochipay.TransactionAction{mochipay.TransactionActionVaultTransfer},
		Status:       mochipay.TransactionStatusSuccess,
		TokenAddress: mochipay.ICYAddress,
		ChainIDs:     []string{mochipay.BASEChainID},
		Size:         math.MaxInt64,
		IsSender:     &isSender,
		SortBy:       "created_at-",
	})
	if err != nil {
		return "", "", err
	}

	earnsICY := big.NewFloat(0)
	earnsUSD := big.NewFloat(0)
	for _, txn := range txns.Data {
		if txn.Metadata["vault_request"] == nil {
			r.logger.Infof("txn %d has no vault_request", txn.Id)
			continue
		}

		vaultRequest := txn.Metadata["vault_request"].(map[string]interface{})
		vaultID, ok := vaultRequest["vault_id"].(float64)
		if !ok {
			r.logger.Infof("vault_id is not int64")
			continue
		}

		if _, ok := dfVaults[int64(vaultID)]; !ok {
			r.logger.Infof("vault_id %d is not in df vaults", int64(vaultID))
			continue
		}

		if txn.CreatedAt.Before(from) || txn.CreatedAt.After(to.Add(24*time.Hour)) {
			r.logger.Infof("txn %d is not in range", txn.Id)
			continue
		}

		if txn.Amount != "" && txn.Token != nil {
			earnsICY.Add(earnsICY, utils.ConvertFromString(txn.Amount, txn.Token.Decimal))
			earnsUSD.Add(earnsUSD, big.NewFloat(txn.UsdAmount))
		}
	}

	return earnsICY.String(), earnsUSD.String(), nil
}
