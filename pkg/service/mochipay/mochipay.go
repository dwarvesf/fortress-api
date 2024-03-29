package mochipay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

const (
	ICYAddress       = "0xf289e3b222dd42b185b7e335fa3c5bd6d132441d"
	BASEChainID      = "8453"
	RewardDefaultMsg = "Send money to treasurer"
	ICYTokenMochiID  = "9d25232e-add3-4bd8-b7c6-be6c14debc58"
	BaseChainName    = "BASE"
)

type IService interface {
	GetListTransactions(req ListTransactionsRequest) (*ListTransactionsResponse, error)
	GetBatchBalances(profileIds []string) (*BatchBalancesResponse, error)
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

	if len(req.Platforms) != 0 {
		platforms := make([]string, 0)
		for _, p := range req.Platforms {
			platforms = append(platforms, string(p))
		}

		queryParams.Add("platforms", strings.Join(platforms, "|"))
	}

	if len(req.ChainIDs) != 0 {
		queryParams.Add("chain_ids", strings.Join(req.ChainIDs, "|"))
	}

	if req.ProfileID != "" {
		queryParams.Add("profile_id", req.ProfileID)
	}

	if req.IsSender != nil {
		queryParams.Add("is_sender", fmt.Sprintf("%v", *req.IsSender))
	}

	if req.SortBy != "" {
		queryParams.Add("sort_by", req.SortBy)
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

func (m *client) GetBatchBalances(profileIds []string) (*BatchBalancesResponse, error) {
	var client = &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/api/v1/mochi-wallet/balances/multiple", m.cfg.MochiPay.BaseURL)
	body, err := json.Marshal(struct {
		ProfileIDs []string `json:"profile_ids"`
	}{ProfileIDs: profileIds})
	if err != nil {
		m.l.Errorf(err, "[mochipay.GetchBatchbalances] json.Marshal failed")
		return nil, err
	}
	r, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		m.l.Errorf(err, "[mochipay.GetBatchBalances] client.Get failed")
		return nil, err
	}
	defer r.Body.Close()

	res := &BatchBalancesResponse{}
	if err := json.NewDecoder(r.Body).Decode(res); err != nil {
		m.l.Errorf(err, "[mochipay.GetBatchBalances] decoder.Decode failed")
		return nil, err
	}

	return res, nil
}
