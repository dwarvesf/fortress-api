package mw

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/metrics"
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

	start := time.Now()
	
	err := amw.authenticate(c)
	if err != nil {
		// Record authentication failure with reason
		method := getAuthMethodFromHeader(c.GetHeader("Authorization"))
		reason := categorizeAuthError(err)
		c.Set("auth_failure_reason", reason)
		
		// Record metrics for authentication failure
		metrics.AuthenticationAttempts.WithLabelValues(method, "failure", reason).Inc()
		
		c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
		return
	}

	// Record successful authentication
	duration := time.Since(start).Seconds()
	method := getAuthMethodFromHeader(c.GetHeader("Authorization"))
	metrics.AuthenticationDuration.WithLabelValues(method).Observe(duration)
	metrics.AuthenticationAttempts.WithLabelValues(method, "success", "").Inc()

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
	claims := &jwt.RegisteredClaims{}

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
		// Record API key validation error
		metrics.APIKeyValidationErrors.WithLabelValues("invalid_format").Inc()
		return ErrInvalidAPIKey
	}

	rec, err := amw.store.APIKey.GetByClientID(amw.repo.DB(), clientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Record API key validation error
			metrics.APIKeyValidationErrors.WithLabelValues("key_not_found").Inc()
			return ErrInvalidAPIKey
		}
		// Record API key validation error
		metrics.APIKeyValidationErrors.WithLabelValues("database_error").Inc()
		return err
	}
	
	if rec.Status != model.ApikeyStatusValid {
		// Record API key validation error
		metrics.APIKeyValidationErrors.WithLabelValues("key_disabled").Inc()
		return ErrInvalidAPIKey
	}

	err = authutils.ValidateHashedKey(rec.SecretKey, key)
	if err != nil {
		// Record API key validation error
		metrics.APIKeyValidationErrors.WithLabelValues("key_mismatch").Inc()
		return err
	}
	
	// Record successful API key usage
	metrics.APIKeyUsage.WithLabelValues(clientID, "success").Inc()
	
	return nil
}

type PermMiddleware struct {
	store  *store.Store
	repo   store.DBRepo
	config *config.Config
}

// WithPerm a middleware to check the permission
func (m *PermMiddleware) WithPerm(perm model.PermissionCode) func(c *gin.Context) {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Set required permission for monitoring
		c.Set("required_permission", perm.String())
		
		accessToken, err := authutils.GetTokenFromRequest(c)
		if err != nil {
			metrics.PermissionChecks.WithLabelValues(perm.String(), "error").Inc()
			c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
			return
		}
		
		tokenType := model.TokenTypeJWT
		if authutils.IsAPIKey(c) {
			tokenType = model.TokenTypeAPIKey
		}

		err = m.ensurePerm(m.store, m.repo.DB(), accessToken, perm.String(), tokenType.String())
		if err != nil {
			// Record permission denial
			duration := time.Since(start).Seconds()
			metrics.PermissionChecks.WithLabelValues(perm.String(), "denied").Inc()
			metrics.AuthorizationDuration.WithLabelValues(perm.String()).Observe(duration)
			
			c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
			return
		}

		// Record successful permission check
		duration := time.Since(start).Seconds()
		metrics.PermissionChecks.WithLabelValues(perm.String(), "allowed").Inc()
		metrics.AuthorizationDuration.WithLabelValues(perm.String()).Observe(duration)

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

// Helper functions for security monitoring
func categorizeAuthError(err error) string {
	switch {
	case errors.Is(err, ErrUnexpectedAuthorizationHeader):
		return "invalid_header_format"
	case errors.Is(err, ErrAuthenticationTypeHeaderInvalid):
		return "invalid_auth_type"
	case errors.Is(err, ErrInvalidAPIKey):
		return "invalid_api_key"
	case strings.Contains(err.Error(), "token is expired"):
		return "token_expired"
	case strings.Contains(err.Error(), "signature is invalid"):
		return "invalid_signature"
	default:
		return "unknown_error"
	}
}

func getAuthMethodFromHeader(authHeader string) string {
	if authHeader == "" {
		return "none"
	}
	
	parts := strings.Split(authHeader, " ")
	if len(parts) < 1 {
		return "unknown"
	}
	
	switch strings.ToLower(parts[0]) {
	case "bearer":
		return "jwt"
	case "apikey":
		return "api_key"
	default:
		return "unknown"
	}
}
