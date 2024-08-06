package reddit

import (
	"context"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

const (
	redditGolangChannel = "golang"
)

// FetchGolangNews fetches the latest Golang news posts from the rising posts.
func (s *service) FetchGolangNews(ctx context.Context) ([]*reddit.Post, error) {
	risingPosts, _, err := s.client.Subreddit.RisingPosts(ctx, redditGolangChannel, &reddit.ListOptions{
		Limit: 50,
	})
	if err != nil {
		return nil, err
	}

	return risingPosts, nil
}
