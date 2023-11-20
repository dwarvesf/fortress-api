package model

const (
	ConfigKeySalaryAdvanceMaxCap = "salary-advance-max-cap"
	ConfigKeyIcyUSDRate          = "icy-usd-rate"
)

type Config struct {
	BaseModel `json:"base_model"`

	Key   string `json:"key"`
	Value string `json:"value"`
}

func (Config) TableName() string { return "configs" }
