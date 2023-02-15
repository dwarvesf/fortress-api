package utils

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	gonanoid "github.com/matoous/go-nanoid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

const (
	alphabet        = "abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNPQRSTUVWXYZ123456789"
	ClientIDLength  = 24
	SecretKeyLength = 32
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

func GenerateUniqueNanoID(length int) (string, error) {
	rs, err := gonanoid.Generate(alphabet, length)
	if err != nil {
		return "", err
	}
	return rs, nil
}

func GenerateHashedKey(key string) (string, error) {
	val := strings.TrimSpace(key)
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(val), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedKey), nil
}

func ExtractAPIKey(apiKey string) (string, string, error) {
	clientID, key := "", ""

	decodedStr, err := base64.StdEncoding.DecodeString(apiKey)
	if err != nil {
		return "", "", err
	}

	decodedAPIKey := string(decodedStr)
	clientID = decodedAPIKey[:ClientIDLength]
	key = decodedAPIKey[ClientIDLength:]

	return clientID, key, nil
}

func ValidateHashedKey(hashedKey string, key string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedKey), []byte(key))
}

func GetUserIDFromToken(cfg *config.Config, tokenString string) (string, error) {
	claims := model.AuthenticationInfo{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecretKey), nil
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

func GetUserIDFromContext(c *gin.Context, cfg *config.Config) (string, error) {
	accessToken, err := GetTokenFromRequest(c)
	if err != nil {
		return "", err
	}

	if IsAPIKey(c) {
		return "", nil
	}

	return GetUserIDFromToken(cfg, accessToken)
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
		return headers[1], nil
	default:
		return "", ErrAuthenticationTypeHeaderInvalid
	}
}

func IsAPIKey(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.Header.Get("Authorization"), "ApiKey")
}

func HasPermission(perms map[string]string, requiredPerm model.PermissionCode) bool {
	_, ok := perms[requiredPerm.String()]

	return ok
}

func GetLoggedInUserInfo(c *gin.Context, storeDB *store.Store, db *gorm.DB, cfg *config.Config) (*model.CurrentLoggedUserInfo, error) {
	if IsAPIKey(c) {
		accessToken, err := GetTokenFromRequest(c)

		clientID, key, err := ExtractAPIKey(accessToken)
		if err != nil {
			return nil, err
		}

		apikey, err := storeDB.APIKey.GetByClientID(db, clientID)
		if err != nil {
			return nil, err
		}

		if apikey.Status != model.ApikeyStatusValid {
			return nil, err
		}

		err = ValidateHashedKey(apikey.SecretKey, key)
		if err != nil {
			return nil, err
		}

		perms, err := storeDB.Permission.GetByApiKeyID(db, apikey.ID.String())
		if err != nil {
			return nil, err
		}

		return &model.CurrentLoggedUserInfo{
			UserID:      clientID,
			Permissions: model.ToPermissionMap(perms),
		}, nil
	}

	userID, err := GetUserIDFromContext(c, cfg)
	if err != nil {
		return nil, err
	}

	e, err := storeDB.Employee.One(db, userID, false)
	if err != nil {
		return nil, err
	}

	perms, err := storeDB.Permission.GetByEmployeeID(db, userID)
	if err != nil {
		return nil, err
	}

	//Get a map of the project and managed flag if they are lead of project.

	projects, err := storeDB.Project.GetByEmployeeID(db, userID)
	if err != nil {
		return nil, err
	}

	projectMap := make(map[model.UUID]*model.Project)
	for _, p := range projects {
		projectMap[p.ID] = p
	}

	rs := &model.CurrentLoggedUserInfo{
		UserID:      userID,
		Permissions: model.ToPermissionMap(perms),
		Projects:    projectMap,
		Role:        e.EmployeeRoles[0].Role.Code,
	}

	return rs, nil
	//return userID, model.ToPermissionMap(perms), projectMap, nil
}
