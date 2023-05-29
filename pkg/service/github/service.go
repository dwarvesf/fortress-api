package github

import (
	"context"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IService interface {
	RemoveFromOrganizationByEmail(ctx context.Context, email string) error
	RemoveFromOrganizationByUsername(ctx context.Context, username string) error
	SendInvitationByEmail(ctx context.Context, e *model.Employee) error
}
