package mochipay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

const (
	ICYAddress       = "0x8D57d71B02d71e1e449a0E459DE40473Eb8f4a90"
	POLYGONChainID   = "137"
	RewardDefaultMsg = "Send money to treasurer"
)

type IService interface {
	GetListTransactions(req ListTransactionsRequest) (*ListTransactionsResponse, error)
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

func (m *client) GetListTransactions(req ListTransactionsRequest) (*ListTransactionsResponse, error) {
	var client = &http.Client{
		Timeout: 10 * time.Second,
	}

	queryParams := url.Values{}
	var pageSize int64 = 10
	if req.Size != 0 {
		pageSize = req.Size
	}
	queryParams.Add("size", fmt.Sprintf("%v", pageSize))
	queryParams.Add("page", fmt.Sprintf("%v", req.Page))

	if len(req.ActionList) > 0 {
		actions := ""
		for i, a := range req.ActionList {
			if i == 0 {
				actions += string(a)
				continue
			}
			actions += fmt.Sprintf("|%s", a)
		}
		queryParams.Add("action", actions)
	}
	if req.Type != "" {
		queryParams.Add("type", string(req.Type))
	}
	if req.Status != "" {
		queryParams.Add("status", string(req.Status))
	}

	if req.TokenAddress != "" {
		queryParams.Add("token_address", req.TokenAddress)
	}

	if req.Participant != "" {
		queryParams.Add("chain_id", req.ChainID)
	}

	if req.Platform != "" {
		queryParams.Add("platform", string(req.Platform))
	}

	if req.ChainID != "" {
		queryParams.Add("chain_id", req.ChainID)
	}

	if req.Participant != "" {
		queryParams.Add("participant", req.Participant)
	}

	url := fmt.Sprintf("%s/api/v1/transactions?", m.cfg.MochiPay.BaseURL) + queryParams.Encode()
	r, err := client.Get(url)
	if err != nil {
		m.l.Errorf(err, "[mochipay.GetListTransaction] client.Get failed")
		return nil, err
	}
	defer r.Body.Close()

	res := &ListTransactionsResponse{}
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		m.l.Errorf(err, "[mochipay.GetListTransaction] decoder.Decode failed")
		return nil, err
	}

	return res, nil
}
