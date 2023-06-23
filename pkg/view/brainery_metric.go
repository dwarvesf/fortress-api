package view

import (
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
	LatestPosts     []Post   `json:"latestPosts"`
	Tags            []string `json:"tags"`
	Contributors    []Post   `json:"contributors"`
	NewContributors []Post   `json:"newContributors"`
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
func ToBraineryMetric(latestPosts, logs []*model.BraineryLog, ncids []string) BraineryMetric {
	metric := BraineryMetric{}

	// latest posts
	for _, post := range latestPosts {
		metric.LatestPosts = append(metric.LatestPosts, ToPost(*post))
	}

	// tags
	tagMap := make(map[string]struct{})
	for _, log := range logs {
		for _, tag := range log.Tags {
			tagMap[tag] = struct{}{}
		}
	}
	for tag := range tagMap {
		metric.Tags = append(metric.Tags, tag)
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

	return metric
}
