package util

import (
	"time"

	"github.com/dgrijalva/jwt-go"

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
