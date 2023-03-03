package employee

import (
	"errors"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type UpdateRoleBody struct {
	Roles []model.UUID
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
	roles, err := r.store.Role.GetByIDs(r.repo.DB(), input.Body.Roles)
	if err != nil {
		return err
	}

	for _, role := range roles {
		if role.Level <= loggedInUser.EmployeeRoles[0].Role.Level &&
			loggedInUser.EmployeeRoles[0].Role.Code != model.AccountRoleAdmin.String() {
			return ErrInvalidAccountRole
		}
	}

	if empl.EmployeeRoles[0].Role.Level == loggedInUser.EmployeeRoles[0].Role.Level &&
		loggedInUser.EmployeeRoles[0].Role.Code != model.AccountRoleAdmin.String() {
		return ErrInvalidAccountRole
	}

	// Begin transaction
	tx, done := r.repo.NewTransaction()

	if err := r.store.EmployeeRole.HardDeleteByEmployeeID(tx.DB(), input.EmployeeID); err != nil {
		return done(err)
	}

	for _, role := range roles {
		_, err = r.store.EmployeeRole.Create(tx.DB(), &model.EmployeeRole{
			EmployeeID: model.MustGetUUIDFromString(input.EmployeeID),
			RoleID:     role.ID,
		})
		if err != nil {
			r.logger.Fields(logger.Fields{
				"emlID":  input.EmployeeID,
				"roleID": role.ID,
			}).Error(err, "failed to create employee role")
			return done(err)
		}
	}

	return done(nil)
}
