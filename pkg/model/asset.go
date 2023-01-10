package model

import "time"

type Asset struct {
	BaseModel

	Name         string
	Price        int64
	Quantity     string
	Note         string
	Location     string
	PurchaseDate *time.Time
}
