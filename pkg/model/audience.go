package model

import "time"

type Audience struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Sources   []string  `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}
