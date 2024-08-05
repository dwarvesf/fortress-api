package news

import (
	"context"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (c *controller) FetchRedditNews(ctx context.Context, tag string) ([]model.News, error) {
	// Fetch Golang news from Reddit
	redditPosts, err := c.service.Reddit.FetchGolangNews(ctx)
	if err != nil {
		return nil, err
	}

	// Truncate to last 10 posts if more than 10
	if len(redditPosts) > 10 {
		redditPosts = redditPosts[len(redditPosts)-10:]
	}

	// Convert to model.News
	news := make([]model.News, len(redditPosts))
	for i, post := range redditPosts {
		news[i] = model.News{
			Title:        post.Title,
			URL:          post.URL,
			Popularity:   int64(post.Score),
			CommentCount: int64(post.NumberOfComments),
			Description:  post.Body,
			Tags:         []string{tag}, // Using the input tag
			CreatedAt:    post.Created.Time,
		}
	}

	return news, nil
}
