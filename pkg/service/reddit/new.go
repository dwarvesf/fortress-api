package reddit

import (
	"fmt"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type service struct {
	client *reddit.Client
}

func New() (IService, error) {
	client, err := reddit.NewReadonlyClient()
	if err != nil {
		return nil, fmt.Errorf("create reddit client failed: %w", err)
	}

	return &service{
		client: client,
	}, nil
}
