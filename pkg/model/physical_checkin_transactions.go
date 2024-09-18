package model

import "time"

type PhysicalCheckinTransaction struct {
	ID         int       `json:"id"`
	EmployeeID UUID      `sql:",type:uuid" json:"employee_id"`
	Date       time.Time `json:"date" gorm:"type:date"`
	IcyAmount  float64   `json:"icy_amount"`
}

func (PhysicalCheckinTransaction) TableName() string { return "physical_checkin_transactions" }
