package mw

import (
	"gorm.io/gorm"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

var noAuthPath = []string{
	"/healthz",
	"/auth",
}

// WithAuth a middleware to check the access token
func WithAuth(c *gin.Context) {
	if !authRequired(c) {
		c.Next()
		return
	}

	err := authenticate(c)
	if err != nil {
		c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
		return
	}

	c.Next()
}

func authRequired(c *gin.Context) bool {
	requestURL := c.Request.URL.Path
	for _, v := range noAuthPath {
		if strings.Contains(requestURL, v) {
			return false
		}
	}

	return true
}

func authenticate(c *gin.Context) error {
	headers := strings.Split(c.Request.Header.Get("Authorization"), " ")
	if len(headers) != 2 {
		return ErrUnexpectedAuthorizationHeader
	}
	switch headers[0] {
	case "Bearer":
		return validateToken(headers[1])
	default:
		return ErrAuthenticationTypeHeaderInvalid
	}
}

// validateToken a func help validate the access token we got
func validateToken(accessToken string) error {
	claims := &jwt.StandardClaims{}

	_, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("JWTSecretKey"), nil
	})
	if err != nil {
		return err
	}

	return claims.Valid()
}
func NewPermissionMiddleware(s *store.Store, r store.DBRepo) *permMiddleware {
	return &permMiddleware{
		store: s,
		repo:  r,
	}
}

type permMiddleware struct {
	store *store.Store
	repo  store.DBRepo
}

// WithPerm a middleware to check the permission
func (m permMiddleware) WithPerm(perm string) func(c *gin.Context) {
	return func(c *gin.Context) {
		accessToken, err := utils.GetTokenFromRequest(c)
		if err != nil {
			c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
			return
		}

		err = ensurePerm(m.store, m.repo.DB(), accessToken, perm)
		if err != nil {
			c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
			return
		}

		c.Next()
	}
}

func ensurePerm(storeDB *store.Store, db *gorm.DB, accessToken string, requiredPerm string) error {
	claims := jwt.MapClaims{}
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("JWTSecretKey"), nil
	})
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return ErrUnauthorized
	}

	userID, ok := claims["id"].(string)
	if !ok {
		return ErrInvalidUserID
	}

	perms, err := storeDB.Permission.GetByEmployeeID(db, userID)
	if err != nil {
		return err
	}

	ok = false
	for _, v := range perms {
		if v.Code == requiredPerm {
			ok = true
			break
		}
	}

	if !ok {
		return errUnauthorized(requiredPerm)
	}

	return nil
}
