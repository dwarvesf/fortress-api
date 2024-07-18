package model

import "time"

type News struct {
	Title      string
	URL        string
	Popularity int64
	CreatedAt  time.Time
}
