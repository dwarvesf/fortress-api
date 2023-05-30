package employee

import (
	"context"
	"errors"
	"github.com/dstotijn/go-notion"
	"github.com/k0kubun/pp/v3"
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

	//now := time.Now()
	e, err := r.store.Employee.One(r.repo.DB(), employeeID, true)
	if err != nil {
		l.Errorf(err, "failed to get Employee ", employeeID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}
	//
	//e.WorkingStatus = body.EmployeeStatus
	//e.LeftDate = &now
	//
	//if body.EmployeeStatus != model.WorkingStatusLeft {
	//	e.LeftDate = nil
	//}
	//
	//tx, done := r.repo.NewTransaction()
	//defer func() {
	//	_ = done(nil)
	//}()
	//
	//_, err = r.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, *e, "working_status", "left_date")
	//if err != nil {
	//	return nil, done(err)
	//}

	//If employee working status is left, do off-boarding flow
	if body.EmployeeStatus == model.WorkingStatusLeft {
		//err = r.store.ProjectMember.UpdateMemberToInActiveByID(tx.DB(), employeeID, &now)
		//if err != nil {
		//	return nil, done(err)
		//}

		// Do Off-boarding process
		r.processOffBoardingEmployee(l, e)
	}

	return e, err
}

func (r *controller) processOffBoardingEmployee(l logger.Logger, e *model.Employee) {
	discordInfo := model.SocialAccounts(e.SocialAccounts).GetDiscord()
	if discordInfo != nil {
		err := r.updateDiscordRoles(discordInfo.AccountID)
		if err != nil {
			l.Errorf(err, "failed to update discord roles", "employeeID", e.ID.String(), "discordID", discordInfo.AccountID)
		}
	}

	err := r.removeBasecampAccess(e.BasecampID)
	if err != nil {
		l.Errorf(err, "failed to remove basecamp access", "employeeID", e.ID.String(), "basecampID", e.BasecampID)
	}

	err = r.removeTeamEmailForward(e.TeamEmail)
	if err != nil {
		l.Errorf(err, "failed to remove team email forward", "employeeID", e.ID.String(), "email", e.TeamEmail)
	}

	err = r.removeTeamEmail(e.TeamEmail)
	if err != nil {
		l.Errorf(err, "failed to delete google account", "employeeID", e.ID.String(), "email", e.TeamEmail)
	}

	err = r.removeGithubFromOrganization(e)
	if err != nil {
		l.Errorf(err, "failed to remove github user from organization", "employeeID", e.ID.String())
	}

	err = r.removeNotionPageAccess(e)
	if err != nil {
		l.Errorf(err, "failed to remove github user from organization", "employeeID", e.ID.String())
	}
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

func (r *controller) removeTeamEmailForward(teamEmail string) error {
	if r.config.Env != "prod" {
		return nil
	}

	err := r.service.ImprovMX.DeleteAccount(teamEmail)
	if err != nil {
		return err
	}

	return nil
}

func (r *controller) removeTeamEmail(teamEmail string) error {
	if r.config.Env != "prod" {
		return nil
	}

	err := r.service.GoogleAdmin.DeleteAccount(teamEmail)
	if err != nil {
		return err
	}

	return nil
}

func (r *controller) removeGithubFromOrganization(e *model.Employee) error {
	if r.config.Env != "prod" {
		return nil
	}

	githubSA := model.SocialAccounts(e.SocialAccounts).GetGithub()
	if githubSA != nil {
		if githubSA.AccountID == "" {
			return nil
		}

		err := r.service.Github.RemoveFromOrganizationByUsername(context.Background(), githubSA.AccountID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *controller) removeNotionPageAccess(e *model.Employee) error {
	//if r.config.Env != "prod" {
	//	return nil
	//}

	rs, err := r.service.Notion.GetPages()
	if err != nil {
		return err
	}

	var pages []notion.Database

	for _, itm := range rs.Results {
		pages = append(pages, itm.(notion.Database))
	}

	p := pages[0]
	pp.Println(p)
	return nil
}
