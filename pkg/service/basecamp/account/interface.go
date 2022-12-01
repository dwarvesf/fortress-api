package account

import (
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type AccountService interface {
	Create(displayName string, email string, org string) (int64, string, error)
	Get(bcID int) (*model.Person, error)
	Remove(userID int64) (err error)
	UpdateInProject(projectID int64, peopleEntry model.PeopleEntry) (id int64, sgID string, err error)
}
