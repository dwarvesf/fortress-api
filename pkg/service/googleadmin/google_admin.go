package googleadmin

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/option"

	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"

	"github.com/dwarvesf/fortress-api/pkg/config"
)

type googleService struct {
	config    *oauth2.Config
	token     *oauth2.Token
	service   *admin.Service
	appConfig *config.Config
}

// New function return Google service
func New(config *oauth2.Config, appConfig *config.Config) IService {
	return &googleService{
		config:    config,
		appConfig: appConfig,
	}
}

func (g *googleService) DeleteAccount(mail string) error {
	if err := g.ensureToken(g.appConfig.Google.AdminGoogleRefreshToken); err != nil {
		return err
	}

	if err := g.prepareService(); err != nil {
		return err
	}

	err := g.service.Users.Delete(mail).Do()
	return err
}

func (g *googleService) GetGroupMemberEmails(groupEmail string) ([]string, error) {
	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return nil, err
	}

	if err := g.prepareService(); err != nil {
		return nil, err
	}

	var memberEmails []string

	members, err := g.service.Members.List(groupEmail).Do()
	if err != nil {
		return nil, err
	}

	if members == nil {
		return nil, fmt.Errorf("No member in group %v", groupEmail)
	}

	for _, m := range members.Members {
		memberEmails = append(memberEmails, m.Email)
	}

	return memberEmails, nil
}

func (g *googleService) ensureToken(rToken string) error {
	token := &oauth2.Token{
		RefreshToken: rToken,
	}

	if !g.token.Valid() {
		tks := g.config.TokenSource(context.Background(), token)
		tok, err := tks.Token()
		if err != nil {
			return err
		}
		g.token = tok
	}
	return nil
}

func (g *googleService) prepareService() error {
	client := g.config.Client(context.Background(), g.token)
	service, err := admin.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return errors.New("failed to prepare google admin service " + err.Error())
	}
	g.service = service
	return nil
}
