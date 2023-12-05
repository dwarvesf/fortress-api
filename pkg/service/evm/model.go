package evm

type RpcClient struct {
	Url string
}

var (
	DefaultPolygonClient = RpcClient{Url: "https://rpc.ankr.com/polygon"}
)
