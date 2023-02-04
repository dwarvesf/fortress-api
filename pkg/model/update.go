package model

import "time"

type Update struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Audience  string    `json:"audience"`
	CreatedAt time.Time `json:"created_at"`
}
