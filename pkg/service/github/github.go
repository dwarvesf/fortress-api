package github

import (
	"context"

	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type githubService struct {
	Client *github.Client
	log    logger.Logger
}

var (
	defaultRole          = "direct_member"
	dwarvesFoundationOrg = "dwarvesf"
)

func New(cfg *config.Config, l logger.Logger) IService {
	if cfg.Github.Token == "" {
		return &githubService{}
	}
	return &githubService{
		Client: github.NewClient(
			oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: cfg.Github.Token,
				},
			))),
		log: l,
	}
}

func (s githubService) SendInvitationByEmail(ctx context.Context, e *model.Employee) error {
	if s.Client == nil {
		s.log.Warn("[SendInvitationByEmail] github token is empty")
		return nil
	}

	role := defaultRole
	opt := github.CreateOrgInvitationOptions{
		Email:  &e.PersonalEmail,
		Role:   &role,
		TeamID: []int64{},
	}

	s.log.Infof("[SendInvitationByEmail] Send invitation to user", "email", e.PersonalEmail)
	_, _, err := s.Client.Organizations.CreateOrgInvitation(ctx, dwarvesFoundationOrg, &opt)
	if err != nil {
		s.log.Errorf(err, "[SendInvitationByEmail] Fail to send invitation", "email", e.PersonalEmail)
		return err
	}

	return nil
}

func (s githubService) RemoveFromOrganizationByEmail(ctx context.Context, email string) error {
	if s.Client == nil {
		return nil
	}

	sOpts := github.SearchOptions{}

	result, _, err := s.Client.Search.Users(ctx, email, &sOpts)
	if err != nil {
		s.log.Errorf(err, "[RemoveFromOrganizationByEmail] fail to search user by email", "email", email)
		return err
	}

	switch {
	case len(result.Users) > 1:
		s.log.Errorf(ErrFoundOneMoreGithubAccount, "[RemoveFromOrganizationByEmail] more than 1 result return", "email", email)
		return ErrFoundOneMoreGithubAccount
	case len(result.Users) == 0:
		s.log.Errorf(ErrFailedToGetGithubAccount, "[RemoveFromOrganizationByEmail] can not found github account from user email", "email", email)
		return ErrFailedToGetGithubAccount
	}

	s.log.Infof("[RemoveFromOrganizationByEmail] Remove github member out of organization", "username", result.Users[0].GetLogin())
	_, err = s.Client.Organizations.RemoveMember(ctx, dwarvesFoundationOrg, result.Users[0].GetLogin())
	if err != nil {
		return err
	}

	return nil
}

func (s githubService) RemoveFromOrganizationByUsername(ctx context.Context, username string) error {
	if s.Client == nil {
		return nil
	}

	result, _, err := s.Client.Users.Get(ctx, username)
	if err != nil {
		s.log.Errorf(err, "[RemoveFromOrganizationByUsername] fail to search user by email", "username", username)
		return err
	}

	s.log.Infof("[RemoveFromOrganizationByUsername] remove github member out of organization", "username", result.GetLogin())
	_, err = s.Client.Organizations.RemoveMember(ctx, dwarvesFoundationOrg, result.GetLogin())
	if err != nil {
		return err
	}

	return nil
}

func (s githubService) RetrieveUsernameByID(ctx context.Context, id int64) (string, error) {
	user, _, err := s.Client.Users.GetByID(ctx, id)
	if err != nil {
		s.log.Errorf(err, "[RetrieveUsernameByID] fail to get user by id", "id", id)
		return "", err
	}

	return user.GetLogin(), nil
}

func (s githubService) FetchOpenPullRequest(ctx context.Context, repo string) (prs []*github.PullRequest, err error) {
	opts := &github.PullRequestListOptions{
		State:     "open",
		Sort:      "created",
		Direction: "desc",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 15,
		},
	}
	prs, _, err = s.Client.PullRequests.List(ctx, dwarvesFoundationOrg, repo, opts)
	if err != nil {
		s.log.Errorf(err, "[FetchOpenPullRequest] fail to fetch pull request", "repo", repo)
		return prs, err
	}

	return prs, nil
}
