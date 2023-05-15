package model

import (
	"time"
)

type IcyTransaction struct {
	BaseModel
	Category       string    `json:"category"`
	TxnTime        time.Time `json:"txn_time"`
	Amount         string    `json:"amount"`
	Note           string    `json:"note"`
	SrcEmployeeId  UUID      `json:"src_employee_id"`
	DestEmployeeId UUID      `json:"dest_employee_id"`
}
