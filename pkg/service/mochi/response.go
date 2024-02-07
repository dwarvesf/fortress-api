package mochi

// GetListVaultsResponse is the response model for mochi-api GetListVaults
type GetListVaultsResponse struct {
	Data []Vault `json:"data"`
}
