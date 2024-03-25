package icyswap

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/dwarvesf/fortress-api/pkg/config"
	icyswapabi "github.com/dwarvesf/fortress-api/pkg/contracts/icyswap"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/evm"
)

type IService interface {
	ConversionRate() (*big.Int, error)
	UsdcFund() (*big.Int, error)
}

type icyswap struct {
	instance *icyswapabi.IcySwap
	evm      evm.IService
	cfg      *config.Config
	l        logger.Logger
}

const (
	ICYSwapAddress = "0x982d2c5A654E4f7CC65ACDCa4ECc649fE4F4DAa4"
	USDCAddress    = "0xd9aAEc86B65D86f6A7B5B1b0c42FFA531710b6CA"
)

func New(evm evm.IService, cfg *config.Config, l logger.Logger) (IService, error) {
	instance, err := icyswapabi.NewIcySwap(common.HexToAddress(ICYSwapAddress), evm.Client())
	if err != nil {
		return nil, err
	}
	return &icyswap{
		instance: instance,
		evm:      evm,
		cfg:      cfg,
		l:        l,
	}, nil
}

func (i *icyswap) ConversionRate() (*big.Int, error) {
	rate, err := i.instance.IcyToUsdcConversionRate(nil)
	if err != nil {
		return nil, err
	}

	return rate, nil
}

func (i *icyswap) UsdcFund() (*big.Int, error) {
	balance, err := i.evm.ERC20Balance(common.HexToAddress(USDCAddress), common.HexToAddress(ICYSwapAddress))
	if err != nil {
		return nil, err
	}
	return balance, nil
}
