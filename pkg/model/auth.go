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
