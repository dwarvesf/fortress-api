package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type News struct {
	Title        string    `json:"title"`
	URL          string    `json:"url"`
	Popularity   int64     `json:"popularity"`
	CommentCount int64     `json:"comment_count"`
	Flag         int64     `json:"flag"`
	Description  string    `json:"description"`
	Tags         []string  `json:"tags"`
	CreatedAt    time.Time `json:"timestamp"`
}

type FetchNewsResponse struct {
	Data []News `json:"data"`
} // @name FetchNewsResponse

func ToFetchNewsResponse(emerging []model.News) []News {
	return ToListNews(emerging)
}

func ToListNews(news []model.News) []News {
	res := make([]News, 0)
	for _, n := range news {
		res = append(res, News{
			Title:        n.Title,
			URL:          n.URL,
			Popularity:   n.Popularity,
			CommentCount: n.CommentCount,
			Flag:         n.Flag,
			Description:  n.Description,
			Tags:         n.Tags,
			CreatedAt:    n.CreatedAt,
		})
	}

	return res
}
