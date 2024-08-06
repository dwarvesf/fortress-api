package reddit

import (
	"context"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type IService interface {
	// FetchGolangNews fetches the latest Golang news posts and filters the rising posts from the new posts.
	FetchGolangNews(ctx context.Context) ([]*reddit.Post, error)
}
