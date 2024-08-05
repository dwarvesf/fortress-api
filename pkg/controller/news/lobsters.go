package news

import (
	"context"
	"sort"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (c *controller) FetchLobstersNews(ctx context.Context, tag string) ([]model.News, error) {
	logger := c.logger.Fields(logger.Fields{
		"controller": "news",
		"method":     "FetchLobstersNews",
		"tag":        tag,
	})

	news, err := c.service.Lobsters.FetchNews(tag)
	if err != nil {
		logger.Error(err, "failed to fetch news from lobsters")
		return nil, err
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
			Title:        n.Title,
			URL:          url,
			Popularity:   int64(n.Score),
			CommentCount: int64(n.CommentCount),
			Description:  n.Description,
			Tags:         n.Tags,
			CreatedAt:    n.CreatedAt,
		})
	}

	// Sort by creation time (descending) and then by popularity (descending)
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].CreatedAt.Equal(normalized[j].CreatedAt) {
			return normalized[i].Popularity > normalized[j].Popularity
		}
		return normalized[i].CreatedAt.After(normalized[j].CreatedAt)
	})

	// Filter for posts within the last 24 hours
	emerging := make([]model.News, 0)
	for _, n := range normalized {
		if time.Since(n.CreatedAt) <= 24*time.Hour {
			emerging = append(emerging, n)
		}
	}

	// If more than 10 posts in 24 hours, truncate to 10
	if len(emerging) > 10 {
		emerging = emerging[:10]
	} else if len(emerging) < 10 {
		// If less than 10 posts in 24 hours, add more posts without time check
		for _, n := range normalized {
			if len(emerging) == 10 {
				break
			}
			if time.Since(n.CreatedAt) > 24*time.Hour {
				emerging = append(emerging, n)
			}
		}
	}

	return emerging, nil
}
