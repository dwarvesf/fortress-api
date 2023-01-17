package utils

import (
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
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

	if accessToken == "ApiKey" {
		return "", nil
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

func IsAPIKey(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.Header.Get("Authorization"), "ApiKey")
}

func HasPermission(c *gin.Context, perms map[string]string, requiredPerm model.PermissionCode) bool {
	if IsAPIKey(c) {
		return true
	}

	_, ok := perms[requiredPerm.String()]

	return ok
}

func GetLoggedInUserInfo(c *gin.Context, storeDB *store.Store, db *gorm.DB) (*model.CurrentLoggedUserInfo, error) {
	if IsAPIKey(c) {
		return &model.CurrentLoggedUserInfo{}, nil
	}

	userID, err := GetUserIDFromContext(c)
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
