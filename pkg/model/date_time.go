package model

import "time"

type DateTime struct {
	Time    time.Time `json:"time"`
	HasTime bool      `json:"has_time"`
}
