package employee

import (
	"errors"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type CreateEmployeeInput struct {
	FullName      string
	DisplayName   string
	TeamEmail     string
	PersonalEmail string
	Positions     []model.UUID
	Salary        int
	SeniorityID   model.UUID
	Roles         []model.UUID
	Status        string
	ReferredBy    model.UUID
}

func (r *controller) Create(userID string, input CreateEmployeeInput) (*model.Employee, error) {
	loggedInUser, err := r.store.Employee.One(r.repo.DB(), userID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEmployeeNotFound
	}

	if err != nil {
		return nil, err
	}

	// 1.2 prepare employee data
	now := time.Now()

	// Check position existence
	positions, err := r.store.Position.All(r.repo.DB())
	if err != nil {
		return nil, err
	}

	positionsReq := make([]model.Position, 0)
	positionMap := model.ToPositionMap(positions)
	for _, pID := range input.Positions {
		_, ok := positionMap[pID]
		if !ok {
			r.logger.Errorf(ErrPositionNotFound, "postion not found with id ", pID.String())
			return nil, ErrPositionNotFound
		}

		positionsReq = append(positionsReq, positionMap[pID])
	}

	sen, err := r.store.Seniority.One(r.repo.DB(), input.SeniorityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSeniorityNotfound
		}
		return nil, err
	}

	roles, err := r.store.Role.GetByIDs(r.repo.DB(), input.Roles)
	if err != nil {
		r.logger.Error(err, "failed to get roles by ids")
		return nil, err
	}

	for _, role := range roles {
		if role.Level <= loggedInUser.EmployeeRoles[0].Role.Level &&
			loggedInUser.EmployeeRoles[0].Role.Code != model.AccountRoleAdmin.String() {
			return nil, ErrInvalidAccountRole
		}
	}

	// get the username
	eml := &model.Employee{
		BaseModel: model.BaseModel{
			ID: model.NewUUID(),
		},
		FullName:      input.FullName,
		DisplayName:   input.DisplayName,
		TeamEmail:     input.TeamEmail,
		PersonalEmail: input.PersonalEmail,
		WorkingStatus: model.WorkingStatus(input.Status),
		JoinedDate:    &now,
		SeniorityID:   sen.ID,
		Username:      strings.Split(input.TeamEmail, "@")[0],
	}

	if !input.ReferredBy.IsZero() {
		exists, err := r.store.Employee.IsExist(r.repo.DB(), input.ReferredBy.String())
		if err != nil {
			return nil, err
		}

		if !exists {
			return nil, ErrReferrerNotFound
		}

		eml.ReferredBy = input.ReferredBy
	}

	// 2.1 check employee exists -> raise error
	_, err = r.store.Employee.OneByEmail(r.repo.DB(), eml.TeamEmail)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			return nil, ErrEmailExisted
		}
		return nil, err
	}

	_, err = r.store.Employee.OneByEmail(r.repo.DB(), eml.PersonalEmail)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			return nil, ErrEmailExisted
		}
		return nil, err
	}

	_, err = r.store.Employee.One(r.repo.DB(), eml.Username, false)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			return nil, ErrEmployeeExisted
		}
		return nil, err
	}

	tx, done := r.repo.NewTransaction()
	// 2.2 store employee
	eml, err = r.store.Employee.Create(tx.DB(), eml)
	if err != nil {
		return nil, done(err)
	}

	// 2.3 create employee position
	for _, p := range positionsReq {
		_, err = r.store.EmployeePosition.Create(tx.DB(), &model.EmployeePosition{
			EmployeeID: eml.ID,
			PositionID: p.ID,
		})
		if err != nil {
			return nil, done(err)
		}
	}

	// 2.4 create employee roles
	for _, role := range roles {
		_, err = r.store.EmployeeRole.Create(tx.DB(), &model.EmployeeRole{
			EmployeeID: eml.ID,
			RoleID:     role.ID,
		})
		if err != nil {
			r.logger.Fields(logger.Fields{
				"emlID":  eml.ID,
				"roleID": role.ID,
			}).Error(err, "failed to create employee role")
			return nil, done(err)
		}
	}

	// Create employee organization
	org, err := r.store.Organization.OneByCode(tx.DB(), model.OrganizationCodeDwarves)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, done(ErrOrganizationNotFound)
		}
		return nil, done(err)
	}

	if _, err := r.store.EmployeeOrganization.Create(tx.DB(), &model.EmployeeOrganization{EmployeeID: eml.ID, OrganizationID: org.ID}); err != nil {
		return nil, done(err)
	}

	return eml, done(nil)
}
