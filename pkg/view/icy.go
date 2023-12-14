package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type IcyAccounting struct {
	ICY                *TokenInfo    `json:"icy"`
	USDT               *TokenInfo    `json:"usdt"`
	IcySwap            *ContractInfo `json:"icySwap"`
	ConversionRate     float32       `json:"conversionRate"`
	CirculatingICY     string        `json:"circulatingICY"`
	ContractFundInUSDT string        `json:"contractFundInUSDT"`
	OffsetUSDT         string        `json:"offsetUSDT"`
}

func ToIcyAccounting(icyAccounting *model.IcyAccounting) *IcyAccounting {
	if icyAccounting == nil {
		return nil
	}
	return &IcyAccounting{
		ICY:                ToTokenInfo(icyAccounting.ICY),
		USDT:               ToTokenInfo(icyAccounting.USDT),
		IcySwap:            ToContractInfo(icyAccounting.IcySwap),
		ConversionRate:     icyAccounting.ConversionRate,
		CirculatingICY:     icyAccounting.CirculatingICY,
		ContractFundInUSDT: icyAccounting.ContractFundInUSDT,
		OffsetUSDT:         icyAccounting.OffsetUSDT,
	}
}

type TokenInfo struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Address     string `json:"address"`
	Decimals    int    `json:"decimals"`
	Chain       string `json:"chain"`
	ChainID     string `json:"chainID"`
	TotalSupply string `json:"totalSupply"`
}

func ToTokenInfo(tokenInfo *model.TokenInfo) *TokenInfo {
	if tokenInfo == nil {
		return nil
	}
	return &TokenInfo{
		Name:        tokenInfo.Name,
		Symbol:      tokenInfo.Symbol,
		Address:     tokenInfo.Address,
		Decimals:    tokenInfo.Decimals,
		Chain:       tokenInfo.Chain,
		ChainID:     tokenInfo.ChainID,
		TotalSupply: tokenInfo.TotalSupply,
	}
}

type ContractInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
}

func ToContractInfo(contractInfo *model.ContractInfo) *ContractInfo {
	if contractInfo == nil {
		return nil
	}
	return &ContractInfo{
		Name:    contractInfo.Name,
		Address: contractInfo.Address,
		Chain:   contractInfo.Chain,
	}
}
