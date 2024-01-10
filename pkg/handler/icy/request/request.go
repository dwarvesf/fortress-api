package request

type CheckICYBalanceRequest struct {
	DiscordID string `json:"discordID"`
} // @name CheckICYBalanceRequest

type WithdrawMoneyRequest struct {
	DiscordID string `json:"discordID"`
	Amount    string `json:"amount"`
} // @name SalaryAdvanceRequest
