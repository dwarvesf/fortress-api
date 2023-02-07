package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (r *controller) Me(c *gin.Context, userID string) (*model.Employee, []*model.Permission, error) {
	e, err := r.store.Employee.One(r.repo.DB(), userID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrUserNotFound
		}

		return nil, nil, err
	}

	perms, err := r.store.Permission.GetByEmployeeID(r.repo.DB(), userID)
	if err != nil {
		return nil, nil, err
	}

	return e, perms, nil
}
