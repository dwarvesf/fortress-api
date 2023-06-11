package mochi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

type IService interface {
	GetVaultTransaction(req *VaultTransactionRequest) (*VaultTransactionResponse, error)
}

type client struct {
	cfg *config.Config
	l   logger.Logger
}

func New(cfg *config.Config, l logger.Logger) IService {
	return &client{
		cfg: cfg,
		l:   l,
	}
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
	resBody, err := ioutil.ReadAll(response.Body)
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
