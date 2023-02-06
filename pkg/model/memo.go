package model

import "time"

type Memo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Tags      []string  `json:"tags"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}
