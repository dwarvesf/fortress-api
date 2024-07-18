package news

import (
	"context"
	"sort"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (c *controller) FetchLobstersNews(ctx context.Context, tag string) ([]model.News, []model.News, error) {
	logger := c.logger.Fields(logger.Fields{
		"controller": "news",
		"method":     "FetchLobstersNews",
		"tag":        tag,
	})

	news, err := c.service.Lobsters.FetchNews(tag)
	if err != nil {
		logger.Error(err, "failed to fetch news from lobsters")
		return nil, nil, err
	}

	normalized := make([]model.News, 0, len(news))
	for _, n := range news {
		url := n.URL
		if url == "" {
			url = n.ShortIDURL
		}
		if url == "" {
			continue
		}

		normalized = append(normalized, model.News{
			Title:      n.Title,
			URL:        url,
			Popularity: int64(n.Score),
			CreatedAt:  n.CreatedAt,
		})
	}

	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].Popularity > normalized[j].Popularity
	})

	var popular, emerging []model.News
	for _, n := range normalized {
		if isIn24Hours(n.CreatedAt) {
			if len(emerging) >= 10 {
				continue
			}
			emerging = append(emerging, n)
		} else {
			if len(popular) >= 10 {
				continue
			}
			popular = append(popular, n)
		}
	}

	sort.Slice(emerging, func(i, j int) bool {
		return emerging[i].CreatedAt.After(emerging[j].CreatedAt)
	})

	return popular, emerging, nil
}

func isIn24Hours(t time.Time) bool {
	return t.After(time.Now().Add(-24 * time.Hour))
}
