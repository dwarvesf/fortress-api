package employee

import (
	"errors"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type UpdateRoleBody struct {
	RoleID model.UUID
}

type UpdateRoleInput struct {
	EmployeeID string
	Body       UpdateRoleBody
}

func (r *controller) UpdateRole(userID string, input UpdateRoleInput) (err error) {
	loggedInUser, err := r.store.Employee.One(r.repo.DB(), userID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrEmployeeNotFound
	}

	if err != nil {
		return err
	}

	empl, err := r.store.Employee.One(r.repo.DB(), input.EmployeeID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrEmployeeNotFound
	}

	if err != nil {
		return err
	}

	// Check role exists
	newRole, err := r.store.Role.One(r.repo.DB(), input.Body.RoleID)
	if err != nil {
		return err
	}

	if empl.EmployeeRoles[0].Role.Level == loggedInUser.EmployeeRoles[0].Role.Level {
		return ErrInvalidAccountRole
	}

	if newRole.Level <= loggedInUser.EmployeeRoles[0].Role.Level {
		return ErrInvalidAccountRole
	}

	// Begin transaction
	tx, done := r.repo.NewTransaction()

	if err := r.store.EmployeeRole.HardDeleteByEmployeeID(tx.DB(), input.EmployeeID); err != nil {
		return done(err)
	}

	_, err = r.store.EmployeeRole.Create(tx.DB(), &model.EmployeeRole{
		EmployeeID: model.MustGetUUIDFromString(input.EmployeeID),
		RoleID:     input.Body.RoleID,
	})
	if err != nil {
		return done(err)
	}

	return done(nil)
}
