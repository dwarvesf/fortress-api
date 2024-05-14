package memologs

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

const (
	dfMemoRssURL = "https://memo.d.foundation/index.xml"
)

// retrieveLatestMemoFeed call to memo.d.foundation to get the rss feed
func retrieveLatestMemoFeed(l logger.Logger) (Rss, error) {
	resp, err := http.Get(dfMemoRssURL)
	if err != nil {
		l.Errorf(err, "failed to get rss feed from %s, status code: %d", dfMemoRssURL, resp.StatusCode)
		return Rss{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error(err, "failed to read rss feed from response body")
		return Rss{}, err
	}

	d := Rss{}
	if err := xml.Unmarshal(data, &d); err != nil {
		l.Errorf(err, "failed to unmarshal rss feed with content: %s", string(data))
	}

	return d, nil
}

func (c *controller) Sync() ([]model.MemoLog, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "memologs",
		"method":     "Sync",
	})

	// TODO: need optimize this, for example just only get synced memos today
	syncedMemos, err := c.store.MemoLog.List(c.repo.DB())
	if err != nil {
		l.Errorf(err, "failed to get synced memos")
		return nil, err
	}

	mapSyncedMemos := make(map[string]model.MemoLog)
	for _, memo := range syncedMemos {
		mapSyncedMemos[memo.URL] = memo
	}

	latestMemoFeed, err := retrieveLatestMemoFeed(l)
	if err != nil {
		return nil, err
	}

	newMemos := make([]model.MemoLog, 0)
	for _, item := range latestMemoFeed.Channel.Item {
		l.Infof("Syncing memo: %s", item.Link)

		// Ignore if memo is already synced
		if _, ok := mapSyncedMemos[item.Link]; ok {
			l.Infof("Memo is already synced: %s", item.Link)
			continue
		}

		// Ignore if memo is not published today
		publishedAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to parse date %v", item.PubDate))
			continue
		}
		if !timeutil.IsSameDay(publishedAt, time.Now()) {
			continue
		}

		// Ignore if memo have no author
		if strings.TrimSpace(item.Author) == "" {
			continue
		}

		// This should be author discord username
		authorUsernames := strings.Split(strings.TrimSpace(item.Author), ",")
		if len(authorUsernames) == 0 {
			continue
		}

		// Get author by memo usernames, ignore any author if not found
		// Expect that we do not have huge number of memo each day, so do not need to fear this loop spam get authors
		authors, err := c.store.DiscordAccount.ListByMemoUsername(c.repo.DB(), authorUsernames)
		if err != nil {
			l.Errorf(err, "failed to get authors by discord usernames: %v", authorUsernames)
			continue
		}

		// Temporary ignore if no author found, do not need to enrich community member info here, it make this function too complex. This task should be done by another cron job
		if len(authors) == 0 {
			continue
		}

		// Create memo log
		memo := model.MemoLog{
			Title:       item.Title,
			URL:         item.Link,
			Description: item.Description,
			PublishedAt: &publishedAt,
			Authors:     authors,
		}

		newMemos = append(newMemos, memo)
	}

	if len(newMemos) == 0 {
		l.Info("No new memo is published today")
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

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Text    string   `xml:",chardata"`
	Atom    string   `xml:"atom,attr"`
	Version string   `xml:"version,attr"`
	Script  string   `xml:"script"`
	Channel struct {
		Text  string `xml:",chardata"`
		Title string `xml:"title"`
		Link  struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"link"`
		Description    string `xml:"description"`
		Generator      string `xml:"generator"`
		Language       string `xml:"language"`
		ManagingEditor string `xml:"managingEditor"`
		WebMaster      string `xml:"webMaster"`
		Copyright      string `xml:"copyright"`
		LastBuildDate  string `xml:"lastBuildDate"`
		Item           []struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			PubDate     string `xml:"pubDate"`
			Author      string `xml:"author"`
			Guid        string `xml:"guid"`
			Description string `xml:"description"`
			Draft       string `xml:"draft"`
		} `xml:"item"`
	} `xml:"channel"`
}
