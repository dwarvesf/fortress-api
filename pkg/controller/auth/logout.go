package auth

import (
	"errors"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (r *controller) Logout(userID string, token string) (*model.Employee, error) {
	em, err := r.store.Employee.One(r.repo.DB(), userID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	tokens, err := r.service.Redis.GetAllBlacklistToken()
	if err != nil {
		return nil, err
	}

	for _, t := range tokens {
		if token == t {
			return nil, ErrInvalidToken
		}
	}

	err = r.service.Redis.AddTokenBlacklist(token)
	if err != nil {
		return nil, err
	}
	return em, nil
}
