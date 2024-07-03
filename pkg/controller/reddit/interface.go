package reddit

import "context"

type IController interface {
	// SyncGolangNews fetches new Golang news from Reddit and sends it to Discord.
	SyncGolangNews(ctx context.Context) error
}
