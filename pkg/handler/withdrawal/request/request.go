package request

type CheckWithdrawConditionRequest struct {
	DiscordID string `json:"discordID" form:"discordID" binding:"required"`
} // @name CheckWithdrawConditionRequest

type PaymentRequestInput struct {
	DiscordID         string `json:"discordID"`
	ICYAmount         string `json:"icyAmount"`
	BankSwiftCode     string `json:"bankSwiftCode"`
	BankAccountNumber string `json:"bankAccountNumber"`
	BankAccountOwner  string `json:"bankAccountOwner"`
} // @name PaymentRequestInput
