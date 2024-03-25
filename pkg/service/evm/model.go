package evm

type RpcClient struct {
	Url string
}

var (
	DefaultBASEClient = RpcClient{Url: "https://mainnet.base.org"}
)
