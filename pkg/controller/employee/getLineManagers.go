package employee

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
)

func (r *controller) GetLineManagers(userInfo *model.CurrentLoggedUserInfo) (employees []*model.Employee, err error) {
	var managers []*model.Employee

	if authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadLineManagerFullAccess) {
		managers, err = r.store.Employee.GetLineManagers(r.repo.DB())
		if err != nil {
			return nil, err
		}
	} else {
		managers, err = r.store.Employee.GetLineManagersOfPeers(r.repo.DB(), userInfo.UserID)
		if err != nil {
			return nil, err
		}
	}

	return managers, nil
}
