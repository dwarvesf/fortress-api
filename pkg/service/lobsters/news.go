package lobsters

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LobsterPost struct {
	ShortID          string    `json:"short_id"`
	ShortIDURL       string    `json:"short_id_url"`
	CreatedAt        time.Time `json:"created_at"`
	Title            string    `json:"title"`
	URL              string    `json:"url"`
	Score            int       `json:"score"`
	Flags            int       `json:"flags"`
	CommentCount     int       `json:"comment_count"`
	Description      string    `json:"description"`
	DescriptionPlain string    `json:"description_plain"`
	CommentsURL      string    `json:"comments_url"`
	SubmitterUser    string    `json:"submitter_user"`
	UserIsAuthor     bool      `json:"user_is_author"`
	Tags             []string  `json:"tags"`
}

const lobstersURL = "https://lobste.rs/"

func (s service) FetchNews(tag string) ([]LobsterPost, error) {
	// Create a new HTTP client
	client := &http.Client{}

	if strings.TrimSpace(tag) == "" {
		return nil, nil
	}

	// Make the GET request
	resp, err := client.Get(lobstersURL + "t/" + tag + ".json")
	if err != nil {
		return nil, fmt.Errorf("make get request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	// Unmarshal the response body
	var posts []LobsterPost
	err = json.Unmarshal(body, &posts)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response body failed: %w", err)
	}

	return posts, nil
}
