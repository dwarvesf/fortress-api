package auth

import (
	"errors"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (c *controller) Me(userID string) (*model.Employee, []*model.Permission, error) {
	e, err := c.store.Employee.One(c.repo.DB(), userID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrUserNotFound
		}

		return nil, nil, err
	}

	perms, err := c.store.Permission.GetByEmployeeID(c.repo.DB(), userID)
	if err != nil {
		return nil, nil, err
	}

	return e, perms, nil
}
