package employee

import (
	"errors"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
)

func (r *controller) Details(id string, userInfo *model.CurrentLoggedUserInfo) (*model.Employee, error) {
	// 2. get employee from store
	rs, err := r.store.Employee.One(r.repo.DB(), id, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	if rs.WorkingStatus == model.WorkingStatusLeft && !authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadFullAccess) {
		return nil, ErrEmployeeNotFound
	}

	mentees, err := r.store.Employee.GetMenteesByID(r.repo.DB(), rs.ID.String())
	if err != nil {
		return nil, err
	}

	if len(mentees) > 0 {
		rs.Mentees = mentees
	}

	return rs, nil
}
