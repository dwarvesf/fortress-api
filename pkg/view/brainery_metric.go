package view

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type Post struct {
	Title       string          `json:"title"`
	URL         string          `json:"url"`
	Reward      decimal.Decimal `json:"reward"`
	PublishedAt string          `json:"publishedAt"`
	DiscordID   string          `json:"discordID"`
}

type BraineryMetric struct {
	LatestPosts     []Post           `json:"latestPosts"`
	Tags            []string         `json:"tags"`
	Contributors    []Post           `json:"contributors"`
	NewContributors []Post           `json:"newContributors"`
	TopContributors []TopContributor `json:"topContributors"`
}

type TopContributor struct {
	Count     int
	Ranking   int
	DiscordID string
}

// ToPost parse BraineryLog model to Post
func ToPost(l model.BraineryLog) Post {
	return Post{
		Title:       l.Title,
		URL:         l.URL,
		Reward:      l.Reward,
		PublishedAt: l.PublishedAt.Format(time.RFC3339Nano),
		DiscordID:   l.DiscordID,
	}
}

// ToBraineryMetric parse BraineryLog logs to BraineryMetric
func ToBraineryMetric(latestPosts, logs []*model.BraineryLog, ncids []string, queryView string) BraineryMetric {
	metric := BraineryMetric{}

	// latest posts
	for _, post := range latestPosts {
		metric.LatestPosts = append(metric.LatestPosts, ToPost(*post))
	}

	// tags
	tagMap := make(map[string]int)
	for _, log := range logs {
		for _, tag := range log.Tags {
			if _, ok := tagMap[tag]; !ok {
				tagMap[tag] = 0
			}
			tagMap[tag]++
		}
	}

	// get 10 top tags
	l := len(tagMap)
	if len(tagMap) > 10 {
		l = 10
	}
	for i := 0; i < l; i++ {
		var topCount int
		var topTag string
		for tag, count := range tagMap {
			if count > topCount {
				topCount = count
				topTag = tag
			}
		}
		metric.Tags = append(metric.Tags, topTag)
		delete(tagMap, topTag)
	}

	ncidsMap := make(map[string]struct{}, len(ncids))
	for _, id := range ncids {
		ncidsMap[id] = struct{}{}
	}

	// contributors
	for _, log := range logs {
		if _, ok := ncidsMap[log.DiscordID]; !ok {
			metric.Contributors = append(metric.Contributors, ToPost(*log))
		}
	}

	// new contributors
	for _, log := range logs {
		if _, ok := ncidsMap[log.DiscordID]; ok {
			metric.NewContributors = append(metric.NewContributors, ToPost(*log))
		}
	}

	// top contributors
	if queryView == "monthly" {
		logMap := make(map[string]int)
		for _, log := range logs {
			if _, ok := logMap[log.DiscordID]; !ok {
				logMap[log.DiscordID] = 0
			}
			logMap[log.DiscordID]++
		}

		for discordID, count := range logMap {
			metric.TopContributors = append(metric.TopContributors, TopContributor{
				Count:     count,
				DiscordID: discordID,
			})
		}

		sort.Slice(metric.TopContributors, func(i, j int) bool {
			return metric.TopContributors[i].Count > metric.TopContributors[j].Count
		})

		if len(metric.TopContributors) > 0 {
			metric.TopContributors[0].Ranking = 1
			for i := 1; i < len(metric.TopContributors); i++ {
				if metric.TopContributors[i].Count == metric.TopContributors[i-1].Count {
					metric.TopContributors[i].Ranking = metric.TopContributors[i-1].Ranking
				} else {
					metric.TopContributors[i].Ranking = metric.TopContributors[i-1].Ranking + 1
				}
			}
		}
	}

	return metric
}
