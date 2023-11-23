package mochi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	mochisdk "github.com/consolelabs/mochi-go-sdk/mochi"
	mochiconfig "github.com/consolelabs/mochi-go-sdk/mochi/config"
	"github.com/consolelabs/mochi-go-sdk/mochi/model"
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

type IService interface {
	GetVaultTransaction(req *VaultTransactionRequest) (*VaultTransactionResponse, error)
	SendFromAccountToUser(amount int, discordID string) ([]model.Transaction, error)
}

type client struct {
	cfg         *config.Config
	l           logger.Logger
	mochiClient mochisdk.APIClient
}

func New(cfg *config.Config, l logger.Logger) IService {
	config := &mochiconfig.Config{
		MochiPay: mochiconfig.MochiPay{
			ApplicationID:   cfg.Mochi.ApplicationID,
			ApplicationName: cfg.Mochi.ApplicationName,
			APIKey:          cfg.Mochi.APIKey,
		},
	}

	mochiClient := mochisdk.NewClient(config)
	return &client{
		cfg:         cfg,
		l:           l,
		mochiClient: mochiClient,
	}
}

// SendFromAccountToUser implements IService.
func (c *client) SendFromAccountToUser(amount int, discordID string) ([]model.Transaction, error) {
	currentMonth := time.Now().Month()
	profile, err := c.mochiClient.GetByDiscordID(discordID)
	if err != nil {
		return nil, err
	}

	txs, err := c.mochiClient.Transfer(&model.TransferRequest{
		RecipientIDs: []string{profile.ID},
		Amounts:      []string{strconv.Itoa(amount)},
		TokenID:      "941f0fb1-00da-49dc-a538-5e81fc874cb4",
		Description:  fmt.Sprintf("%s Addvance Salary in %s", discordID, currentMonth.String()),
		References:   "Advance Salary",
	})
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func (m *client) GetVaultTransaction(req *VaultTransactionRequest) (*VaultTransactionResponse, error) {
	var client = &http.Client{}
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/vault/%s/transaction?start_time=%s&end_time=%s", m.cfg.Mochi.BaseURL, req.VaultID, req.StartTime, req.EndTime), nil)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	res := &VaultTransactionResponse{}
	err = json.Unmarshal(resBody, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
