package model

import (
	"time"
)

type HiringPosition struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Projects  []string  `json:"project"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
