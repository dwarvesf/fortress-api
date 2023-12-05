package icy

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/icyswap"
)

type IController interface {
	Accounting() (*model.IcyAccounting, error)
}

type controller struct {
	service *service.Service
	logger  logger.Logger
	config  *config.Config
}

func New(service *service.Service, logger logger.Logger, cfg *config.Config) IController {
	return &controller{
		service: service,
		logger:  logger,
		config:  cfg,
	}
}

func (c *controller) Accounting() (*model.IcyAccounting, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "icy",
		"method":     "Accounting",
	})

	// 0.Prepare token and swap contract data
	icy := c.icy()
	usdt := c.usdt()
	icySwap := c.icySwap()
	icyAddress := common.HexToAddress(icy.Address)

	// 1.Get current conversion rate from icyswap contract
	conversionRate, err := c.service.IcySwap.ConversionRate()
	if err != nil {
		l.Error(err, "failed to get icy conversion rate")
		return nil, err
	}
	usdtDecimals := new(big.Float).SetInt(math.BigPow(10, int64(usdt.Decimals)))
	conversionRateFloat, _ := new(big.Float).Quo(new(big.Float).SetInt(conversionRate), usdtDecimals).Float32()

	// 2. Get current usdt fund in icyswap contract
	icyswapUsdtBal, err := c.service.IcySwap.UsdtFund()
	if err != nil {
		l.Error(err, "failed to get usdt fund in icyswap contract")
		return nil, err
	}

	// 3. Get Circulating Icy
	// circulating icy = total supply - icy in contract - icy of team - icy in mochi app - icy in vault

	// 3.1 Get total icy supply
	icyTotalSupply, _ := new(big.Int).SetString(icy.TotalSupply, 10)

	// 3.2 Get total locked icy amount
	lockedIcyAmount, err := c.getLockedIcy()
	if err != nil {
		c.logger.Error(err, "failed to get locked icy amount")
		return nil, err
	}

	// 3.3 Calculate circulating icy amount
	circulatingIcy := new(big.Int).Sub(icyTotalSupply, lockedIcyAmount)

	// 4.Get offset usdt
	// offset usd: circulating icy in usdt - usd fund in contract -> get how many usd left to redeem

	circulatingIcyInUsdt := new(big.Int).Mul(circulatingIcy, conversionRate)
	// continue to divide to 10^18 for get the amount in usdt decimals
	circulatingIcyInUsdt = new(big.Int).Div(circulatingIcyInUsdt, math.BigPow(10, 18))

	offsetUsdt := new(big.Int).Sub(circulatingIcyInUsdt, icyswapUsdtBal)

	// 5. Return accounting result
	accounting := &model.IcyAccounting{
		Icy:                icy,
		Usdt:               usdt,
		IcySwap:            icySwap,
		ConversionRate:     conversionRateFloat,
		ContractFundInUsdt: icyswapUsdtBal.String(),
		CirculatingIcy:     circulatingIcy.String(),
		OffsetUSDT:         offsetUsdt.String(),
	}

	return accounting, nil
}

func (c *controller) getLockedIcy() (*big.Int, error) {
	icyAddress := common.HexToAddress(c.icy().Address)
	oldIcySwapContractAddr := common.HexToAddress("0xd327b6d878bcd9d5ec6a5bc99445985d75f0d6e5")
	icyswapAddr := common.HexToAddress(icyswap.IcySwapAddress)
	teamAddr := common.HexToAddress("0x0762c4b40c9cb21Af95192a3Dc3EDd3043CF3d41")
	icyLockedAddrs := []common.Address{oldIcySwapContractAddr, icyswapAddr, teamAddr}

	type FetchResult struct {
		amount *big.Int
		err    error
	}
	fetchIcyResults := make(chan FetchResult)

	wg := sync.WaitGroup{}
	wg.Add(len(icyLockedAddrs))
	go func() {
		wg.Wait()
		close(fetchIcyResults)
	}()

	for _, lockedAddr := range icyLockedAddrs {
		go func(ownerAddr common.Address) {
			amount, err := c.service.PolygonClient.ERC20Balance(icyAddress, ownerAddr)
			fetchIcyResults <- FetchResult{
				amount: amount,
				err:    err,
			}
			wg.Done()
		}(lockedAddr)
	}

	lockedIcyAmount := big.NewInt(0)
	for res := range fetchIcyResults {
		if res.err != nil {
			return nil, res.err
		}
		lockedIcyAmount = new(big.Int).Add(lockedIcyAmount, res.amount)
	}

	vaultIcyBal, _ := new(big.Int).SetString("1000000000000000000", 10)
	lockedIcyAmount = new(big.Int).Add(lockedIcyAmount, vaultIcyBal)

	return lockedIcyAmount, nil
}

func (c *controller) icy() model.TokenInfo {
	return model.TokenInfo{
		Name:        "Icy",
		Symbol:      "ICY",
		Address:     "0x8D57d71B02d71e1e449a0E459DE40473Eb8f4a90",
		Decimals:    18,
		Chain:       "Polygon",
		TotalSupply: "100000000000000000000000",
	}
}

func (c *controller) usdt() model.TokenInfo {
	return model.TokenInfo{
		Name:     "Usdt",
		Symbol:   "USDT",
		Address:  "0xc2132D05D31c914a87C6611C10748AEb04B58e8F",
		Decimals: 6,
		Chain:    "Polygon",
	}
}

func (c *controller) icySwap() model.ContractInfo {
	return model.ContractInfo{
		Name:    "IcySwap",
		Address: icyswap.IcySwapAddress,
		Chain:   "Polygon",
	}
}
