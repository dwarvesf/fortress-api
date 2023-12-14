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
	UsdtFund() (*big.Int, error)
}

type icyswap struct {
	instance *icyswapabi.IcySwap
	evm      evm.IService
	cfg      *config.Config
	l        logger.Logger
}

const (
	IcySwapAddress = "0x8De345A73625237223dEDf8c93dfE79A999C17FB"
	UsdtAddress    = "0xc2132D05D31c914a87C6611C10748AEb04B58e8F"
)

func New(evm evm.IService, cfg *config.Config, l logger.Logger) (IService, error) {
	instance, err := icyswapabi.NewIcySwap(common.HexToAddress(IcySwapAddress), evm.Client())
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

func (i *icyswap) UsdtFund() (*big.Int, error) {
	balance, err := i.evm.ERC20Balance(common.HexToAddress(UsdtAddress), common.HexToAddress(IcySwapAddress))
	if err != nil {
		return nil, err
	}
	return balance, nil
}
