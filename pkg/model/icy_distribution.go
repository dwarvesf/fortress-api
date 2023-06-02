package model

// IcyDistribution is a model for icy_distribution table
type IcyDistribution struct {
	BaseModel
	Team   string `json:"team"`
	Period string `json:"period"`
	Amount string `json:"amount"`
}
