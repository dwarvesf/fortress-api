package news

import (
	"context"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IController interface {
	FetchLobstersNews(ctx context.Context, tag string) ([]model.News, error)
	FetchRedditNews(ctx context.Context, tag string) ([]model.News, error)
}
