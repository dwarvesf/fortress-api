package utils

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetTokenFromRequest(c *gin.Context) (string, error) {
	headers := strings.Split(c.Request.Header.Get("Authorization"), " ")
	if len(headers) != 2 {
		return "", errors.New("unexpected headers")
	}
	switch headers[0] {
	case "Bearer":
		return headers[1], nil
	default:
		return "", errors.New("authentication type is invalid")
	}
}
