package request

type CheckWithdrawConditionRequest struct {
	DiscordID string `json:"discordID" form:"discordID" binding:"required"`
} // @name CheckWithdrawConditionRequest

type WithdrawMoneyRequest struct {
	DiscordID string `json:"discordID"`
	Amount    string `json:"amount"`
} // @name WithdrawMoneyRequest
