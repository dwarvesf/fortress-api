package github

import (
	"context"

	"github.com/google/go-github/v52/github"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IService interface {
	RemoveFromOrganizationByEmail(ctx context.Context, email string) error
	RemoveFromOrganizationByUsername(ctx context.Context, username string) error
	SendInvitationByEmail(ctx context.Context, e *model.Employee) error
	RetrieveUsernameByID(ctx context.Context, id int64) (string, error)
	FetchOpenPullRequest(ctx context.Context, repo string) ([]*github.PullRequest, error)
}
