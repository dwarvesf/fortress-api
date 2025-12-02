package employee

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
)

type CreateEmployeeInput struct {
	FullName      string
	DisplayName   string
	TeamEmail     string
	PersonalEmail string
	Positions     []model.UUID
	Salary        int64
	SeniorityID   model.UUID
	Roles         []model.UUID
	Status        string
	ReferredBy    model.UUID
	JoinDate      *time.Time
	SkipEmail     bool
}

func (r *controller) Create(userID string, input CreateEmployeeInput) (*model.Employee, error) {
	l := r.logger.Fields(logger.Fields{
		"controller": "employee",
		"method":     "Create",
	})

	loggedInUser, err := r.store.Employee.One(r.repo.DB(), userID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEmployeeNotFound
	}
	if err != nil {
		return nil, err
	}

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
			l.Errorf(ErrPositionNotFound, "position not found with id ", pID.String())
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
		l.Error(err, "failed to get roles by ids")
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
		JoinedDate:    input.JoinDate,
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
			return nil, ErrTeamEmailExisted
		}
		return nil, err
	}

	_, err = r.store.Employee.OneByEmail(r.repo.DB(), eml.PersonalEmail)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			return nil, ErrPersonalEmailExisted
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
		l.Errorf(err, "failed to create employee", "employee", eml)
		return nil, done(err)
	}

	// 2.3 create employee position
	for _, p := range positionsReq {
		ep := &model.EmployeePosition{
			EmployeeID: eml.ID,
			PositionID: p.ID,
		}
		_, err = r.store.EmployeePosition.Create(tx.DB(), ep)
		if err != nil {
			l.Errorf(err, "failed to create employee position", "employee_position", ep)
			return nil, done(err)
		}
	}

	// 2.4 create employee roles
	for _, role := range roles {
		er := &model.EmployeeRole{
			EmployeeID: eml.ID,
			RoleID:     role.ID,
		}
		_, err = r.store.EmployeeRole.Create(tx.DB(), &model.EmployeeRole{
			EmployeeID: eml.ID,
			RoleID:     role.ID,
		})
		if err != nil {
			l.Errorf(err, "failed to create employee role", "employee_role", er)
			return nil, done(err)
		}
	}

	baseCurrency, err := r.store.Currency.GetByName(tx.DB(), currency.VNDCurrency)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, done(ErrCurrencyNotFound)
		}

		return nil, done(err)
	}

	salaryBatch := 1
	if input.JoinDate.Day() > 1 && input.JoinDate.Day() < 16 {
		salaryBatch = 15
	}

	// 2.4 create employee salary
	ebs := &model.BaseSalary{
		EmployeeID:            eml.ID,
		ContractAmount:        0,
		CompanyAccountAmount:  0,
		PersonalAccountAmount: input.Salary,
		InsuranceAmount:       0,
		Type:                  "",
		Category:              "",
		CurrencyID:            baseCurrency.ID,
		Batch:                 salaryBatch,
		EffectiveDate:         nil,
	}
	err = r.store.BaseSalary.Save(tx.DB(), ebs)
	if err != nil {
		l.Errorf(err, "failed to create employee base salary", "employee_base_salary", ebs)
		return nil, done(err)
	}

	// Create employee organization
	org, err := r.store.Organization.OneByCode(tx.DB(), model.OrganizationCodeDwarves)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, done(ErrOrganizationNotFound)
		}
		return nil, done(err)
	}

	eo := &model.EmployeeOrganization{
		EmployeeID:     eml.ID,
		OrganizationID: org.ID,
	}
	if _, err := r.store.EmployeeOrganization.Create(tx.DB(), eo); err != nil {
		l.Errorf(err, "failed to create employee organization", "employee_organization", eo)
		return nil, done(err)
	}

	authenticationInfo := model.AuthenticationInfo{
		UserID: eml.ID.String(),
		Avatar: eml.Avatar,
		Email:  eml.PersonalEmail,
	}

	jwt, err := authutils.GenerateJWTToken(&authenticationInfo, time.Now().Add(24*time.Hour).Unix(), r.config.JWTSecretKey)
	if err != nil {
		l.Errorf(err, "failed to generate jwt token", "authenticationInfo", authenticationInfo)
		return nil, done(err)
	}

	ei := model.EmployeeInvitation{
		EmployeeID:               eml.ID,
		InvitedBy:                loggedInUser.ID,
		InvitationCode:           jwt,
		IsCompleted:              false,
		IsInfoUpdated:            false,
		IsDiscordRoleAssigned:    false,
		IsBasecampAccountCreated: false,
		IsTeamEmailCreated:       false,
	}

	if _, err := r.store.EmployeeInvitation.Create(tx.DB(), &ei); err != nil {
		l.Errorf(err, "failed to create employee invitation", "employee_invitation", ei)
		return nil, done(err)
	}

	if !input.SkipEmail {
		l.Debug("sending invitation email to employee")
		invitation := model.InvitationEmail{
			Email:   eml.PersonalEmail,
			Link:    fmt.Sprintf("%s/onboarding?code=%s", r.config.FortressURL, jwt),
			Inviter: loggedInUser.FullName,
		}

		if err := r.service.GoogleMail.SendInvitationMail(&invitation); err != nil {
			l.Errorf(err, "failed to send invitation mail", "invitationInfo", invitation)
			return nil, done(err)
		}
	} else {
		l.Debug("skipping invitation email as requested")
	}

	return eml, done(nil)
}
