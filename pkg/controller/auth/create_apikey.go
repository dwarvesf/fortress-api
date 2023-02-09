package auth

import (
	"encoding/base64"
	"errors"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/gin-gonic/gin"
)

func (r *controller) CreateAPIkey(c *gin.Context, roleID string) (string, error) {
	clientID, err := utils.GenerateUniqueNanoID(utils.ClientIDLength)
	if err != nil {
		return "", err
	}
	key, err := utils.GenerateUniqueNanoID(utils.SecretKeyLength)
	if err != nil {
		return "", err
	}

	hashedKey, err := utils.GenerateHashedKey(key)
	if err != nil {
		return "", err
	}

	roleIDUUID, err := model.UUIDFromString(roleID)
	if err != nil {
		return "", err
	}

	tx, done := r.repo.NewTransaction()

	role, err := r.store.Role.One(tx.DB(), roleIDUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", done(ErrRoleNotfound)
		}
		return "", done(err)
	}

	apikey, err := r.store.APIkey.Create(tx.DB(), &model.Apikey{
		ClientID:  clientID,
		SecretKey: hashedKey,
		Status:    model.ApikeyStatusValid,
	})
	if err != nil {
		return "", done(err)
	}

	_, err = r.store.ApikeyRole.Create(tx.DB(), &model.ApikeyRole{
		ApiKeyID: apikey.ID,
		RoleID:   role.ID,
	})
	if err != nil {
		return "", done(err)
	}

	return base64.StdEncoding.EncodeToString([]byte(clientID + key)), done(nil)
}