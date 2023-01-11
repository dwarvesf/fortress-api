package model

import (
	"github.com/dgrijalva/jwt-go"
)

// AuthenticationInfo ..
type AuthenticationInfo struct {
	jwt.StandardClaims

	UserID string `json:"id"`
	Avatar string `json:"avatar"`
	Email  string `json:"email"`
}

type CurrentLoggedUserInfo struct {
	UserID      string
	Permissions map[string]string
	Projects    map[UUID]*Project
	Role        string
}
