package people

import (
	"errors"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

var (
	ErrNotInProject = errors.New("this account does not belong to this project")
)

type Service interface {
	GetByID(id int) (res *model.Person, err error)
	GetInfo() (res *model.UserInfo, err error)
	Create(name string, email string, orgnization string) (id int64, sgID string, err error)
	Remove(userID int64) (err error)
	UpdateInProject(projectID int64, peopleEntry model.PeopleEntry) (id int64, sgID string, err error)
	GetAllOnProject(projectID int) (result []model.Person, err error)
}
