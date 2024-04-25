package icy

import (
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/icyswap"
	"github.com/dwarvesf/fortress-api/pkg/service/mochipay"
	"github.com/dwarvesf/fortress-api/pkg/service/mochiprofile"
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
	usdt := c.usdc()
	icySwap := c.icySwap()

	// 1.Get current conversion rate from icyswap contract
	conversionRate, err := c.service.IcySwap.ConversionRate()
	if err != nil {
		l.Error(err, "failed to get icy conversion rate")
		return nil, err
	}
	usdtDecimals := new(big.Float).SetInt(math.BigPow(10, int64(usdt.Decimals)))
	conversionRateFloat, _ := new(big.Float).Quo(new(big.Float).SetInt(conversionRate), usdtDecimals).Float32()

	// 2. Get current usdt fund in icyswap contract
	icyswapUsdtBal, err := c.service.IcySwap.UsdcFund()
	if err != nil {
		l.Error(err, "failed to get usdt fund in icyswap contract")
		return nil, err
	}

	// 3. Get Circulating Icy
	// circulating icy = total supply - icy in contract - icy of team - icy in mochi app - icy in vault

	// 3.1 Get total icy supply
	icyTotalSupply, _ := new(big.Int).SetString(icy.TotalSupply, 10)

	// 3.2 Get total locked icy amount
	lockedIcyAmount, err := c.lockedIcyAmount()
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
	return &model.IcyAccounting{
		ICY:                &icy,
		USDT:               &usdt,
		IcySwap:            &icySwap,
		ConversionRate:     conversionRateFloat,
		ContractFundInUSDT: icyswapUsdtBal.String(),
		CirculatingICY:     circulatingIcy.String(),
		OffsetUSDT:         offsetUsdt.String(),
	}, nil
}

func (c *controller) lockedIcyAmount() (*big.Int, error) {
	lockedIcyAmount := big.NewInt(0)

	// 0. fetch onchain locked icy amount
	onchainLockedAmount, err := c.onchainLockedIcyAmount()
	if err != nil {
		c.logger.Error(err, "failed to get onchain locked icy amount")
		return nil, err
	}
	lockedIcyAmount = new(big.Int).Add(lockedIcyAmount, onchainLockedAmount)

	// 1. fetch offchain locked icy amount
	offchainLockedAmount, err := c.offchainLockedIcyAmount()
	if err != nil {
		c.logger.Error(err, "failed to get offchain locked icy amount")
		return nil, err
	}
	lockedIcyAmount = new(big.Int).Add(lockedIcyAmount, offchainLockedAmount)

	// 2. return result
	return lockedIcyAmount, nil
}

func (c *controller) onchainLockedIcyAmount() (*big.Int, error) {
	icyAddress := common.HexToAddress(c.icy().Address)
	oldIcySwapContractAddr := common.HexToAddress("0xd327b6d878bcd9d5ec6a5bc99445985d75f0d6e5")
	icyswapAddr := common.HexToAddress(icyswap.ICYSwapAddress)
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
			amount, err := c.service.BaseClient.ERC20Balance(icyAddress, ownerAddr)
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

	return lockedIcyAmount, nil
}

func (c *controller) offchainLockedIcyAmount() (*big.Int, error) {
	// 0. get all profile, which type is vault or app
	profileIds := make([]string, 0)
	const pageSize int64 = 50
	var page int64 = 0
	for {
		res, err := c.service.MochiProfile.GetListProfiles(mochiprofile.ListProfilesRequest{
			Types: []mochiprofile.ProfileType{
				mochiprofile.ProfileTypeApplication,
				mochiprofile.ProfileTypeVault,
			},
			Page: page,
			Size: pageSize,
		})
		if err != nil {
			return nil, err
		}
		for _, p := range res.Data {
			profileIds = append(profileIds, p.ID)
		}

		hasNext := res.Pagination.Total/pageSize-page > 0
		if !hasNext {
			break
		}
		page += 1
	}

	// 1. get balance of all profiles
	balRes, err := c.service.MochiPay.GetBatchBalances(profileIds)
	if err != nil {
		return nil, err
	}

	total := big.NewInt(0)
	icy := c.icy()
	for _, b := range balRes.Data {
		// filter token icy
		if strings.EqualFold(b.Token.Address, icy.Address) && strings.EqualFold(b.Token.ChainId, icy.ChainID) {
			amount, _ := new(big.Int).SetString(b.Amount, 10)
			total = new(big.Int).Add(total, amount)
		}
	}

	return total, nil
}

func (c *controller) icy() model.TokenInfo {
	return model.TokenInfo{
		Name:        "Icy",
		Symbol:      "ICY",
		Address:     mochipay.ICYAddress,
		Decimals:    18,
		Chain:       mochipay.BASEChainID,
		ChainID:     mochipay.BASEChainID,
		TotalSupply: "100000000000000000000000",
	}
}

func (c *controller) usdc() model.TokenInfo {
	return model.TokenInfo{
		Name:     "USD Base Coin",
		Symbol:   "USDbC",
		Address:  icyswap.USDCAddress,
		Decimals: 6,
		Chain:    mochipay.BaseChainName,
		ChainID:  mochipay.BASEChainID,
	}
}

func (c *controller) icySwap() model.ContractInfo {
	return model.ContractInfo{
		Name:    "IcySwap",
		Address: icyswap.ICYSwapAddress,
		Chain:   mochipay.BaseChainName,
	}
}
