package employee

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/utils"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type UpdateWorkingStatusInput struct {
	EmployeeStatus model.WorkingStatus
}

func (r *controller) UpdateEmployeeStatus(employeeID string, body UpdateWorkingStatusInput) (*model.Employee, error) {
	l := r.logger.Fields(logger.Fields{
		"controller": "employee",
		"method":     "UpdateEmployeeStatus",
	})

	now := time.Now()
	emp, err := r.store.Employee.One(r.repo.DB(), employeeID, true)
	if err != nil {
		l.Errorf(err, "failed to get Employee ", employeeID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	emp.WorkingStatus = body.EmployeeStatus
	emp.LeftDate = &now

	if body.EmployeeStatus != model.WorkingStatusLeft {
		emp.LeftDate = nil
	}

	tx, done := r.repo.NewTransaction()
	defer func() {
		_ = done(nil)
	}()

	_, err = r.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, *emp, "working_status", "left_date")
	if err != nil {
		return nil, done(err)
	}

	// If employee working status is left, do off-boarding flow
	if body.EmployeeStatus == model.WorkingStatusLeft {
		err = r.store.ProjectMember.UpdateMemberToInActiveByID(tx.DB(), employeeID, &now)
		if err != nil {
			return nil, done(err)
		}

		discordInfo := model.SocialAccounts(emp.SocialAccounts).GetDiscord()
		if discordInfo != nil {
			err = r.updateDiscordRoles(discordInfo.AccountID)
			if err != nil {
				l.Errorf(err, "failed to update discord roles", "employeeID", employeeID, "discordID", discordInfo.AccountID)
				return nil, err
			}
		}

		err = r.removeBasecampAccess(emp.BasecampID)
		if err != nil {
			l.Errorf(err, "failed to remove basecamp access", "employeeID", employeeID, "basecampID", emp.BasecampID)
			return nil, err
		}
	}

	return emp, err
}

func (r *controller) updateDiscordRoles(discordUserID string) error {
	if r.config.Env != "prod" {
		return nil
	}

	if discordUserID == "" {
		return nil
	}

	roles, err := r.service.Discord.GetRoles()
	if err != nil {
		return err
	}

	dfRoles := roles.DwarvesRoles()

	discordMember, err := r.service.Discord.GetMember(discordUserID)
	if err != nil {
		return err
	}

	for _, role := range dfRoles {
		if utils.Contains(discordMember.Roles, role.ID) {
			err = r.service.Discord.RemoveRole(discordUserID, role.ID)
			if err != nil {
				return err
			}
		}
	}

	// Assign alumni role
	alumniRole := roles.ByCode("alumni")
	err = r.service.Discord.AddRole(discordUserID, alumniRole.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *controller) removeBasecampAccess(baseCampID int) error {
	if r.config.Env != "prod" {
		return nil
	}

	err := r.service.Basecamp.People.Remove(int64(baseCampID))
	if err != nil {
		return err
	}

	return nil
}
