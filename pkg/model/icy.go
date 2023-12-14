package model

type IcyAccounting struct {
	ICY     *TokenInfo    `json:"icy"`
	USDT    *TokenInfo    `json:"usdt"`
	IcySwap *ContractInfo `json:"icy_swap"`

	ConversionRate     float32 `json:"conversion_rate"`
	CirculatingICY     string  `json:"circulating_icy"`
	ContractFundInUSDT string  `json:"contract_fund_in_usdt"`
	OffsetUSDT         string  `json:"offset_usdt"` // how many usdt left need to be issued
}

type TokenInfo struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Address     string `json:"address"`
	Decimals    int    `json:"decimals"`
	Chain       string `json:"chain"`
	ChainID     string `json:"chain_id"`
	TotalSupply string `json:"total_supply"`
}

type ContractInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Chain   string `json:"chain"`
}
