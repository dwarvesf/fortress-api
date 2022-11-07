package accountstatus

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All() (accountStatuses []*model.AccountStatus, err error)
}
