package model

type IcyAccounting struct {
	Icy     TokenInfo    `json:"icy"`
	Usdt    TokenInfo    `json:"usdt"`
	IcySwap ContractInfo `json:"icySwap"`

	ConversionRate     float32 `json:"conversionRate"`
	CirculatingIcy     string  `json:"circulatingIcy"`
	ContractFundInUsdt string  `json:"contractFundInUsdt"`
	OffsetUSDT         string  `json:"offsetUsdt"` // how many usdt left need to be issued
}

type TokenInfo struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Address     string `json:"address"`
	Decimals    int    `json:"decimals"`
	Chain       string `json:"chain"`
	ChainID     string `json:"chain_id"`
	TotalSupply string `json:"totalSupply"`
}

type ContractInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
}
