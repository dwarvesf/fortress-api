package model

import (
	"time"
)

type NotionEvent struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Date         DateTime  `json:"date"`
	ActivityType string    `json:"activity_type"`
	CreatedAt    time.Time `json:"created_at"`
}
