package improvmx

import (
	"errors"
	"strings"

	"github.com/issyl0/go-improvmx"
)

type improvMXService struct {
	service *improvmx.Client
}

// New function return Google service
func New(token string) IService {
	return &improvMXService{
		service: improvmx.NewClient(token),
	}
}

func (g *improvMXService) CreateAccount(email, emailFwd string) error {
	alias := strings.Replace(email, "@d.foundation", "", -1)

	resp := g.service.CreateEmailForward("d.foundation", alias, emailFwd)
	if !resp.Success {
		return errors.New(resp.Error)
	}

	return nil
}

func (g *improvMXService) DeleteAccount(mail string) error {
	alias := strings.Replace(mail, "@d.foundation", "", -1)

	resp := g.service.DeleteEmailForward("d.foundation", alias)
	if !resp.Success {
		return errors.New(resp.Error)
	}
	return nil
}
