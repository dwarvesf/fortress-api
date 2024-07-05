package reddit

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type service struct {
	client *reddit.Client
}

// Initialize a custom HTTP client
var defaultClient = http.Client{
	Transport: &http.Transport{
		// This will disable http/2
		TLSClientConfig: &tls.Config{},
	},
	Timeout: 10 * time.Second, // Adjust the timeout as needed
}

func New() (IService, error) {
	client, err := reddit.NewReadonlyClient(
		reddit.WithHTTPClient(&defaultClient),
	)
	if err != nil {
		return nil, fmt.Errorf("create reddit client failed: %w", err)
	}

	return &service{
		client: client,
	}, nil
}
