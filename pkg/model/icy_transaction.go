package model

import (
	"time"
)

type IcyTransaction struct {
	BaseModel

	Category       string
	TxnTime        time.Time
	Amount         string
	Note           string
	SrcEmployeeID  UUID
	DestEmployeeID UUID
	Sender         string
	Target         string
}
