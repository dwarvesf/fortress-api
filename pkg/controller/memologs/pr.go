package memologs

import (
	"context"
	"fmt"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

var (
	repoMemos            = []string{"brainery", "playground", "playbook"}
	dwarvesFoundationOrg = "dwarvesf"
)

func (c controller) ListOpenPullRequest() (map[string][]model.MemoPullRequest, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "memologs",
		"method":     "ListOpenPullRequest",
	})

	memoPrs := make(map[string][]model.MemoPullRequest, len(repoMemos))

	// Get discord accounts
	discordAccounts, err := c.store.DiscordAccount.All(c.repo.DB())
	if err != nil {
		l.Error(err, "failed to get discord accounts")
		return nil, fmt.Errorf("failed to get discord accounts: %w", err)
	}

	mapGithubWithDiscord := make(map[string]string, len(discordAccounts))
	for _, dc := range discordAccounts {
		if dc != nil && dc.GithubUsername != "" {
			mapGithubWithDiscord[dc.GithubUsername] = dc.DiscordID
		}
	}

	for _, repo := range repoMemos {
		memoRepoName := fmt.Sprintf("%s/%s", dwarvesFoundationOrg, repo)

		prs, err := c.service.Github.FetchOpenPullRequest(context.Background(), repo)
		if err != nil {
			l.Errorf(err, "failed to fetch open pull requests for %s", repo)
			return memoPrs, err
		}

		memoPullRequest := make([]model.MemoPullRequest, 0, len(prs))
		for _, pr := range prs {
			user := pr.GetUser()
			githubUserName := ""

			if user != nil {
				githubUserName = user.GetLogin()
			}

			discordID := mapGithubWithDiscord[githubUserName]

			prMap := model.MemoPullRequest{
				Number:         pr.GetNumber(),
				Title:          pr.GetTitle(),
				Url:            pr.GetHTMLURL(),
				DiscordId:      discordID,
				GithubUserName: githubUserName,
				Timestamp:      pr.GetCreatedAt().Time,
			}
			memoPullRequest = append(memoPullRequest, prMap)
		}

		memoPrs[memoRepoName] = memoPullRequest
	}

	return memoPrs, nil
}
