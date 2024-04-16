package auth

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
)

type AuthenticationInput struct {
	Code        string
	RedirectURL string
}

func (c *controller) Auth(in AuthenticationInput) (*model.Employee, string, error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "auth",
		"method":     "Auth",
	})

	accessToken, err := c.service.Google.GetAccessToken(in.Code, in.RedirectURL)
	if err != nil {
		l.Errorf(err, "failed to get access token")
		return nil, "", err
	}

	// 2.2 get login user email from access token
	primaryEmail := ""
	if c.config.Env == "prod" {
		primaryEmail, err = c.service.Google.GetGoogleEmailLegacy(accessToken)
		if err != nil {
			l.Errorf(err, "failed to get google email legacy")
			return nil, "", err
		}
	} else {
		primaryEmail, err = c.service.Google.GetGoogleEmail(accessToken)
		if err != nil {
			l.Errorf(err, "failed to get google email")
			return nil, "", err
		}
	}

	// 2.3 double check empty primary email
	if primaryEmail == "" {
		return nil, "", ErrEmptyPrimaryEmail
	}

	// 2.4 check user is active
	employee, err := c.store.Employee.OneByEmail(c.repo.DB(), primaryEmail)
	if err != nil {
		l.Errorf(err, "failed to employee by email")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrUserInactivated
		}
		return nil, "", err
	}

	if employee.WorkingStatus == model.WorkingStatusLeft || employee.WorkingStatus == model.WorkingStatusOnBoarding {
		return nil, "", ErrUserInactivated
	}

	// 2.5 generate jwt bearer token
	authenticationInfo := model.AuthenticationInfo{
		UserID: employee.ID.String(),
		Avatar: employee.Avatar,
		Email:  primaryEmail,
	}

	jwt, err := authutils.GenerateJWTToken(&authenticationInfo, time.Now().Add(24*365*time.Hour).Unix(), c.config.JWTSecretKey)
	if err != nil {
		l.Errorf(err, "failed to generate jwt token")
		return nil, "", err
	}

	return employee, jwt, nil
}
