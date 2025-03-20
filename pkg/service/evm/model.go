package evm

type RpcClient struct {
	Url string
}

var (
	DefaultBASEClient   = RpcClient{Url: "https://mainnet.base.org"}
	TestnetBASEClient   = RpcClient{Url: "https://chain-proxy.wallet.coinbase.com?targetName=base-sepolia"}
	TestnetBASEExplorer = "https://sepolia.basescan.org"
)
