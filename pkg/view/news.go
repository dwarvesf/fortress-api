package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type News struct {
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	Popularity int64     `json:"popularity"`
	CreatedAt  time.Time `json:"timestamp"`
}

type ListNews struct {
	Popular  []News `json:"popular"`
	Emerging []News `json:"emerging"`
}

type FetchNewsResponse struct {
	Data ListNews `json:"data"`
} // @name FetchNewsResponse

func ToFetchNewsResponse(popular, emerging []model.News) ListNews {
	return ListNews{
		Popular:  ToListNews(popular),
		Emerging: ToListNews(emerging),
	}
}

func ToListNews(news []model.News) []News {
	res := make([]News, 0)
	for _, n := range news {
		res = append(res, News{
			Title:      n.Title,
			URL:        n.URL,
			Popularity: n.Popularity,
			CreatedAt:  n.CreatedAt,
		})
	}

	return res
}
