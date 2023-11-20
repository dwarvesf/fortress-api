package view

type CheckSalaryAdvance struct {
	AmountIcy string `json:"amount_icy"`
	AmountUsd string `json:"amount_usd"`
}

func ToCheckSalaryAdvance(amountIcy, amountUSD string) *CheckSalaryAdvance {
	return &CheckSalaryAdvance{
		AmountIcy: amountIcy,
		AmountUsd: amountUSD,
	}
}

type SalaryAdvance struct {
	AmountIcy       string `json:"amount_icy"`
	AmountUsd       string `json:"amount_usd"`
	TransactionID   string `json:"transaction_id"`
	TransactionHash string `json:"transaction_hash"`
}

func ToSalaryAdvance(amountIcy, amountUSD, transactionID, transactionHash string) *SalaryAdvance {
	return &SalaryAdvance{
		AmountIcy:       amountIcy,
		AmountUsd:       amountUSD,
		TransactionID:   transactionID,
		TransactionHash: transactionHash,
	}
}
