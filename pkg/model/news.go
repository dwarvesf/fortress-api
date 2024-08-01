package model

import "time"

type News struct {
	Title        string
	URL          string
	Popularity   int64
	CommentCount int64
	Flag         int64
	Description  string
	Tags         []string
	CreatedAt    time.Time
}
