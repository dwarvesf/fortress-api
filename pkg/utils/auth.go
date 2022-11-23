package utils

import (
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// GenerateJWTToken ...
func GenerateJWTToken(info *model.AuthenticationInfo, expiresAt int64, secretKey string) (string, error) {
	info.ExpiresAt = expiresAt
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, info)
	encryptedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return encryptedToken, nil
}

func GetUserIDFromToken(tokenString string) (string, error) {
	claims := model.AuthenticationInfo{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("JWTSecretKey"), nil
	})

	if !token.Valid {
		return "", ErrInvalidToken
	}
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return "", ErrInvalidSignature
		}
		return "", ErrBadToken
	}
	if time.Unix(claims.ExpiresAt, 0).Before(time.Now()) {
		return "", ErrInvalidToken
	}
	return claims.UserID, nil
}

func GetUserIDFromContext(c *gin.Context) (string, error) {
	accessToken, err := GetTokenFromRequest(c)
	if err != nil {
		return "", err
	}

	return GetUserIDFromToken(accessToken)
}

func GetTokenFromRequest(c *gin.Context) (string, error) {
	headers := strings.Split(c.Request.Header.Get("Authorization"), " ")
	if len(headers) != 2 {
		return "", ErrUnexpectedAuthorizationHeader
	}
	switch headers[0] {
	case "Bearer":
		return headers[1], nil
	case "ApiKey":
		return "ApiKey", nil
	default:
		return "", ErrAuthenticationTypeHeaderInvalid
	}
}
