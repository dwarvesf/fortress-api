package view

import "github.com/dwarvesf/fortress-api/pkg/model"

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

type SalaryAdvanceReportResponse struct {
	PaginationResponse
	Data SalaryAdvanceReport `json:"data"`
} // @name SalaryAdvanceReportResponse

type AggregatedSalaryAdvance struct {
	EmployeeID      string  `json:"employeeID"`
	DiscordID       string  `json:"discordID"`
	DiscordUsername string  `json:"discordUsername"`
	AmountICY       int64   `json:"amountICY"`
	AmountUSD       float64 `json:"amountUSD"`
} // @name AggregatedSalaryAdvance

type SalaryAdvanceReport struct {
	SalaryAdvances []AggregatedSalaryAdvance `json:"salaryAdvances"`
	TotalICY       int64                     `json:"totalICY"`
	TotalUSD       float64                   `json:"totalUSD"`
} // @name SalaryAdvanceReport

func ToSalaryAdvanceReport(result model.SalaryAdvanceReport) *SalaryAdvanceReport {
	var salaryAdvances []AggregatedSalaryAdvance
	for _, v := range result.SalaryAdvances {
		salaryAdvances = append(salaryAdvances, AggregatedSalaryAdvance{
			EmployeeID:      v.EmployeeID,
			DiscordID:       v.Employee.DiscordAccount.DiscordID,
			DiscordUsername: v.Employee.DiscordAccount.Username,
			AmountICY:       v.AmountICY,
			AmountUSD:       v.AmountUSD,
		})
	}

	return &SalaryAdvanceReport{
		SalaryAdvances: salaryAdvances,
		TotalICY:       result.TotalICY,
		TotalUSD:       result.TotalUSD,
	}
}
