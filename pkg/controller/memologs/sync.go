package memologs

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store/memolog"
)

const (
	dfMemoRssURL = "https://memo.d.foundation/index.xml"
)

func (c *controller) Sync() ([]model.MemoLog, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "memologs",
		"method":     "Sync",
	})

	last7days := time.Now().AddDate(0, 0, -7)
	latestMemos, err := c.store.MemoLog.List(c.repo.DB(), memolog.ListFilter{})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorf(err, "failed to get latest memo")
		return nil, err
	}

	latestMemosMap := make(map[string]model.MemoLog)
	for _, memo := range latestMemos {
		latestMemosMap[memo.URL] = memo
	}

	resp, err := http.Get(dfMemoRssURL)
	if err != nil {
		l.Errorf(err, "failed to get rss feed from %s, status code: %d", dfMemoRssURL, resp.StatusCode)
		return nil, err
	}
	defer resp.Body.Close()

	decoder := xml.NewDecoder(resp.Body)

	inItem := false
	currentElem := ""
	item := Item{}
	newMemos := make([]model.MemoLog, 0)
	stop := false
	for !stop {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			currentElem = se.Name.Local
			if se.Name.Local == "item" {
				inItem = true
				item = Item{}
			}
		case xml.EndElement:
			if se.Name.Local == "item" {
				inItem = false

				if item.Link == "" {
					continue
				}

				if _, ok := latestMemosMap[item.Link]; ok {
					continue
				}

				title := strings.TrimSpace(item.Title)
				title = strings.TrimPrefix(title, "<![CDATA[")
				title = strings.TrimSuffix(title, "]]>")
				title = strings.TrimSpace(title)
				if title == "null" || title == "" {
					// Extract from link - just get the final part
					if item.Link != "" {
						parts := strings.Split(item.Link, "/")
						if len(parts) > 0 {
							lastPart := parts[len(parts)-1]
							// Convert kebab-case to Title Case
							words := strings.Split(lastPart, "-")
							for i, word := range words {
								words[i] = cases.Title(language.English).String(word)
							}
							title = strings.Join(words, " ")
						}
					}
				}

				pubDateStr := strings.TrimSpace(item.PubDate)
				pubDate, err := time.Parse(time.RFC1123, pubDateStr)
				if err != nil {
					pubDate = time.Now()
				}
				if pubDate.Before(last7days) {
					stop = true
					break
				}

				description := strings.TrimSpace(item.Description)
				description = strings.TrimPrefix(description, "<![CDATA[")
				description = strings.TrimSuffix(description, "]]>")
				description = strings.TrimSpace(description)

				content := strings.TrimSpace(item.Content)
				content = strings.TrimPrefix(content, "<![CDATA[")
				content = strings.TrimSuffix(content, "]]>")
				content = strings.TrimSpace(content)

				// Use content if description is empty or null
				if description == "" || description == "null" {
					if content != "" {
						// Try to extract first paragraph from content as description
						lines := strings.Split(content, "\n")
						for _, line := range lines {
							line = strings.TrimSpace(line)
							if line != "" && !strings.HasPrefix(line, "#") {
								description = line
								break
							}
						}
					}
				}

				// Clean up CDATA sections and handle null values
				creator := strings.TrimSpace(item.Creator)
				creator = strings.TrimPrefix(creator, "<![CDATA[")
				creator = strings.TrimSuffix(creator, "]]>")
				creator = strings.TrimSpace(creator)

				newMemos = append(newMemos, model.MemoLog{
					Title:             title,
					URL:               item.Link,
					Description:       description,
					PublishedAt:       &pubDate,
					DiscordAccountIDs: []string{creator},
					Category:          extractMemoCategory(item.Link),
					Tags:              extractMemoCategory(item.Link),
				})
			}
		case xml.CharData:
			if inItem {
				data := strings.TrimSpace(string(se))
				if data == "" {
					continue
				}
				switch currentElem {
				case "title":
					item.Title = data
				case "link":
					item.Link = data
				case "pubDate":
					item.PubDate = data
				case "creator":
					item.Creator = data
				case "description":
					item.Description = data
				case "encoded":
					item.Content = data
				}
			}
		}
	}

	if len(newMemos) == 0 {
		return nil, nil
	}

	// Create new memos
	results, err := c.store.MemoLog.Create(c.repo.DB(), newMemos)
	if err != nil {
		c.logger.Errorf(err, "failed to create new memos")
		return nil, err
	}

	return results, nil
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Creator     string `xml:"dc:creator"`
	Guid        string `xml:"guid"`
	Description string `xml:"description"`
	Content     string `xml:"content:encoded"`
}

// extractMemoCategory extracts memo category from link
func extractMemoCategory(url string) []string {
	routes := strings.Split(url, "memo.d.foundation")
	if len(routes) < 2 {
		return nil
	}

	splitPath := strings.Split(strings.TrimSpace(routes[1]), "/")

	// Filter out empty string
	category := make([]string, 0)
	for _, s := range splitPath {
		if s != "" {
			category = append(category, s)
		}
	}

	// final thing is the name of memo, ignore it
	if len(category) > 0 {
		category = category[:len(category)-1]
	}

	return category
}
