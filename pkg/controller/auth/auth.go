package auth

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

type AuthenticationInput struct {
	Code        string
	RedirectURL string
}

func (r *controller) Auth(c *gin.Context, in AuthenticationInput) (*model.Employee, string, error) {
	// 2.1 get access token from req code and redirect url
	accessToken, err := r.service.Google.GetAccessToken(in.Code, in.RedirectURL)
	if err != nil {
		return nil, "", err
	}

	// 2.2 get login user email from access token
	primaryEmail, err := r.service.Google.GetGoogleEmail(accessToken)
	if err != nil {
		return nil, "", err
	}

	// 2.3 double check empty primary email
	if primaryEmail == "" {
		return nil, "", err
	}

	// 2.4 check user is active
	employee, err := r.store.Employee.OneByEmail(r.repo.DB(), primaryEmail)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrUserInactivated
		}
		return nil, "", err
	}

	// 2.5 generate jwt bearer token
	authenticationInfo := model.AuthenticationInfo{
		UserID: employee.ID.String(),
		Avatar: employee.Avatar,
		Email:  primaryEmail,
	}

	jwt, err := utils.GenerateJWTToken(&authenticationInfo, time.Now().Add(24*365*time.Hour).Unix(), r.config.JWTSecretKey)
	if err != nil {
		return nil, "", err
	}

	return employee, jwt, nil
}
