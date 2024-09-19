package model

import "time"

type PhysicalCheckinTransaction struct {
	ID         UUID
	EmployeeID UUID
	Date       time.Time
	IcyAmount  float64
	MochiTxID  int64
}

func (PhysicalCheckinTransaction) TableName() string { return "physical_checkin_transactions" }
