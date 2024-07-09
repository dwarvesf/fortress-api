package memologs

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strings"
	"time"

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
	latestMemos, err := c.store.MemoLog.List(c.repo.DB(), memolog.ListFilter{
		From: &last7days,
	})
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

				if _, ok := latestMemosMap[item.Link]; ok {
					continue
				}

				// Ignore folder is built as an item
				if item.Description == "" && item.Author == "" {
					continue
				}

				pubDate, _ := time.Parse(time.RFC1123Z, item.PubDate)
				if pubDate.Before(last7days) {
					stop = true
					break
				}

				authorUsernames := make([]string, 0)
				for _, s := range strings.Split(strings.TrimSpace(item.Author), ",") {
					if s != "" {
						authorUsernames = append(authorUsernames, strings.TrimSpace(s))
					}
				}

				authors, err := c.store.DiscordAccount.ListByMemoUsername(c.repo.DB(), authorUsernames)
				if err != nil {
					l.Errorf(err, "failed to get authors by discord usernames: %v", authorUsernames)
					continue
				}

				newMemos = append(newMemos, model.MemoLog{
					Title:               item.Title,
					URL:                 item.Link,
					Description:         item.Description,
					PublishedAt:         &pubDate,
					Authors:             authors,
					AuthorMemoUsernames: authorUsernames,
					Category:            extractMemoCategory(item.Link),
				})
			}
		case xml.CharData:
			if inItem {
				data := string(se)
				switch currentElem {
				case "title":
					item.Title = data
				case "link":
					item.Link = data
				case "pubDate":
					item.PubDate = data
				case "author":
					item.Author = data
				case "guid":
					item.Guid = data
				case "description":
					item.Description = data
				case "draft":
					item.Draft = data
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
		l.Errorf(err, "failed to create new memos")
		return nil, err
	}

	return results, nil
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Author      string `xml:"author"`
	Guid        string `xml:"guid"`
	Description string `xml:"description"`
	Draft       string `xml:"draft"`
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
