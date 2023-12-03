package view

type CheckSalaryAdvance struct {
	AmountICY string `json:"amountICY"`
	AmountUSD string `json:"amountUSD"`
} // @name CheckSalaryAdvance

type CheckSalaryAdvanceResponse struct {
	Data CheckSalaryAdvance `json:"data"`
} // @name CheckSalaryAdvanceResponse

func ToCheckSalaryAdvance(amountIcy, amountUSD string) *CheckSalaryAdvance {
	return &CheckSalaryAdvance{
		AmountICY: amountIcy,
		AmountUSD: amountUSD,
	}
}

type SalaryAdvance struct {
	AmountICY       string `json:"amountICY"`
	AmountUSD       string `json:"amountUSD"`
	TransactionID   string `json:"transactionID"`
	TransactionHash string `json:"transactionHash"`
} // @name SalaryAdvance

type SalaryAdvanceResponse struct {
	Data SalaryAdvance `json:"data"`
} // @name SalaryAdvanceResponse

func ToSalaryAdvance(amountIcy, amountUSD, transactionID, transactionHash string) *SalaryAdvance {
	return &SalaryAdvance{
		AmountICY:       amountIcy,
		AmountUSD:       amountUSD,
		TransactionID:   transactionID,
		TransactionHash: transactionHash,
	}
}
