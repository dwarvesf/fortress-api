package auth

import (
	"encoding/base64"
	"errors"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
)

func (r *controller) CreateAPIKey(roleID string) (string, error) {
	clientID, err := authutils.GenerateUniqueNanoID(authutils.ClientIDLength)
	if err != nil {
		return "", err
	}
	key, err := authutils.GenerateUniqueNanoID(authutils.SecretKeyLength)
	if err != nil {
		return "", err
	}

	hashedKey, err := authutils.GenerateHashedKey(key)
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

	apikey, err := r.store.APIKey.Create(tx.DB(), &model.APIKey{
		ClientID:  clientID,
		SecretKey: hashedKey,
		Status:    model.ApikeyStatusValid,
	})
	if err != nil {
		return "", done(err)
	}

	_, err = r.store.APIKeyRole.Create(tx.DB(), &model.APIKeyRole{
		APIKeyID: apikey.ID,
		RoleID:   role.ID,
	})
	if err != nil {
		return "", done(err)
	}

	return base64.URLEncoding.EncodeToString([]byte(clientID + key)), done(nil)
}
