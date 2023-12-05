package evm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/contracts/erc20"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

type IService interface {
	Client() *ethclient.Client
	ERC20Balance(address, owner common.Address) (*big.Int, error)
}

type evm struct {
	client *ethclient.Client
	cfg    *config.Config
	l      logger.Logger
}

func New(rpc RpcClient, cfg *config.Config, l logger.Logger) (IService, error) {
	client, err := ethclient.Dial(rpc.Url)
	if err != nil {
		return nil, err
	}

	return &evm{
		client: client,
		cfg:    cfg,
		l:      l,
	}, nil
}

func (e *evm) Client() *ethclient.Client {
	return e.client
}

func (e *evm) ERC20Balance(address, owner common.Address) (*big.Int, error) {
	instance, err := erc20.NewErc20(address, e.client)
	if err != nil {
		return nil, err
	}

	balance, err := instance.BalanceOf(nil, owner)
	if err != nil {
		return nil, err
	}

	return balance, nil
}
