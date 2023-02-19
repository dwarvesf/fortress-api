package mw

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
)

var noAuthPath = []string{
	"/healthz",
}

type AuthMiddleware struct {
	cfg   *config.Config
	store *store.Store
	repo  store.DBRepo
}

func NewAuthMiddleware(cfg *config.Config, s *store.Store, r store.DBRepo) *AuthMiddleware {
	return &AuthMiddleware{
		cfg:   cfg,
		store: s,
		repo:  r,
	}
}

// WithAuth a middleware to check the access token
func (amw *AuthMiddleware) WithAuth(c *gin.Context) {
	if !authRequired(c) {
		c.Next()
		return
	}

	err := amw.authenticate(c)
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

func (amw *AuthMiddleware) authenticate(c *gin.Context) error {
	headers := strings.Split(c.Request.Header.Get("Authorization"), " ")
	if len(headers) != 2 {
		return ErrUnexpectedAuthorizationHeader
	}
	switch headers[0] {
	case "Bearer":
		return amw.validateToken(headers[1])
	case "ApiKey":
		return amw.validateAPIKey(headers[1])
	default:
		return ErrAuthenticationTypeHeaderInvalid
	}
}

// validateToken a func help validate the access token we got
func (amw *AuthMiddleware) validateToken(accessToken string) error {
	claims := &jwt.StandardClaims{}

	_, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(amw.cfg.JWTSecretKey), nil
	})
	if err != nil {
		return err
	}

	return claims.Valid()
}
func NewPermissionMiddleware(s *store.Store, r store.DBRepo, cfg *config.Config) *PermMiddleware {
	return &PermMiddleware{
		store:  s,
		repo:   r,
		config: cfg,
	}
}

func (amw *AuthMiddleware) validateAPIKey(apiKey string) error {
	clientID, key, err := authutils.ExtractAPIKey(apiKey)
	if err != nil {
		return ErrInvalidAPIKey
	}

	rec, err := amw.store.APIKey.GetByClientID(amw.repo.DB(), clientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidAPIKey
		}
		return err
	}
	if rec.Status != model.ApikeyStatusValid {
		return ErrInvalidAPIKey
	}

	return authutils.ValidateHashedKey(rec.SecretKey, key)
}

type PermMiddleware struct {
	store  *store.Store
	repo   store.DBRepo
	config *config.Config
}

// WithPerm a middleware to check the permission
func (m *PermMiddleware) WithPerm(perm model.PermissionCode) func(c *gin.Context) {
	return func(c *gin.Context) {
		accessToken, err := authutils.GetTokenFromRequest(c)
		if err != nil {
			c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
			return
		}
		tokenType := model.TokenTypeJWT
		if authutils.IsAPIKey(c) {
			tokenType = model.TokenTypeAPIKey
		}

		err = m.ensurePerm(m.store, m.repo.DB(), accessToken, perm.String(), tokenType.String())
		if err != nil {
			c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
			return
		}

		c.Next()
	}
}

func (m *PermMiddleware) ensurePerm(storeDB *store.Store, db *gorm.DB, accessToken string, requiredPerm string, tokenType string) error {
	var perms []*model.Permission

	if tokenType == model.TokenTypeAPIKey.String() {
		clientID, key, err := authutils.ExtractAPIKey(accessToken)
		if err != nil {
			return ErrInvalidAPIKey
		}

		apikey, err := storeDB.APIKey.GetByClientID(db, clientID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvalidAPIKey
			}
			return err
		}

		if apikey.Status != model.ApikeyStatusValid {
			return ErrInvalidAPIKey
		}

		err = authutils.ValidateHashedKey(apikey.SecretKey, key)
		if err != nil {
			return err
		}

		perms, err = storeDB.Permission.GetByApiKeyID(db, apikey.ID.String())
		if err != nil {
			return err
		}
	}

	if tokenType == model.TokenTypeJWT.String() {
		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.config.JWTSecretKey), nil
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

		perms, err = storeDB.Permission.GetByEmployeeID(db, userID)
		if err != nil {
			return err
		}
	}

	ok := false
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
