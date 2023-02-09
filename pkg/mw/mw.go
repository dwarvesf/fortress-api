package mw

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

var noAuthPath = []string{
	"/healthz",
}

type authMiddleware struct {
	cfg   *config.Config
	store *store.Store
	repo  store.DBRepo
}

func NewAuthMiddleware(cfg *config.Config, s *store.Store, r store.DBRepo) *authMiddleware {
	return &authMiddleware{
		cfg:   cfg,
		store: s,
		repo:  r,
	}
}

// WithAuth a middleware to check the access token
func (amw *authMiddleware) WithAuth(c *gin.Context) {
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

func (mw *authMiddleware) authenticate(c *gin.Context) error {
	headers := strings.Split(c.Request.Header.Get("Authorization"), " ")
	if len(headers) != 2 {
		return ErrUnexpectedAuthorizationHeader
	}
	switch headers[0] {
	case "Bearer":
		return mw.validateToken(headers[1])
	case "ApiKey":
		return mw.validateApikey(headers[1])
	default:
		return ErrAuthenticationTypeHeaderInvalid
	}
}

// validateToken a func help validate the access token we got
func (mw *authMiddleware) validateToken(accessToken string) error {
	claims := &jwt.StandardClaims{}

	_, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(mw.cfg.JWTSecretKey), nil
	})
	if err != nil {
		return err
	}

	return claims.Valid()
}
func NewPermissionMiddleware(s *store.Store, r store.DBRepo, cfg *config.Config) *permMiddleware {
	return &permMiddleware{
		store:  s,
		repo:   r,
		config: cfg,
	}
}

func (mw *authMiddleware) validateApikey(apiKey string) error {
	clientID, key, err := utils.ExtractAPIKey(apiKey)
	if err != nil {
		return ErrInvalidAPIKey
	}

	rec, err := mw.store.APIkey.GetByClientID(mw.repo.DB(), clientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidAPIKey
		}
		return err
	}
	if rec.Status != model.ApikeyStatusValid {
		return ErrInvalidAPIKey
	}

	return utils.ValidateHashedKey(rec.SecretKey, key)
}

type permMiddleware struct {
	store  *store.Store
	repo   store.DBRepo
	config *config.Config
}

// WithPerm a middleware to check the permission
func (m permMiddleware) WithPerm(perm model.PermissionCode) func(c *gin.Context) {
	return func(c *gin.Context) {
		accessToken, err := utils.GetTokenFromRequest(c)
		if err != nil {
			c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
			return
		}
		tokenType := "JWT"
		if utils.IsAPIKey(c) {
			tokenType = "ApiKey"
		}

		err = m.ensurePerm(m.store, m.repo.DB(), accessToken, perm.String(), tokenType)
		if err != nil {
			c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
			return
		}

		c.Next()
	}
}

func (m *permMiddleware) ensurePerm(storeDB *store.Store, db *gorm.DB, accessToken string, requiredPerm string, tokenType string) error {
	var perms []*model.Permission
	if tokenType == "ApiKey" {
		clientID, key, err := utils.ExtractAPIKey(accessToken)
		if err != nil {
			return ErrInvalidAPIKey
		}
		apikey, err := storeDB.APIkey.GetByClientID(db, clientID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvalidAPIKey
			}
			return err
		}
		if apikey.Status != model.ApikeyStatusValid {
			return ErrInvalidAPIKey
		}

		err = utils.ValidateHashedKey(apikey.SecretKey, key)
		if err != nil {
			return err
		}

		perms, err = storeDB.Permission.GetByApiKeyID(db, apikey.ID.String())
		if err != nil {
			return err
		}
	} else {
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
