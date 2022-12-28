package workunitmember

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, wum *model.WorkUnitMember) error
	GetByWorkUnitID(db *gorm.DB, wuID string) (wuMembers []*model.WorkUnitMember, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.WorkUnitMember, updatedFields ...string) (*model.WorkUnitMember, error)
	DeleteByWorkUnitID(db *gorm.DB, workUnitID string) error
	All(db *gorm.DB, workUnitID string) (members []*model.WorkUnitMember, err error)
	One(db *gorm.DB, workUnitID string, employeeID string, status string) (workUnitMember *model.WorkUnitMember, err error)
	SoftDeleteByWorkUnitID(db *gorm.DB, workUnitID string, employeeID string) (err error)
	GetPeerReviewerInTimeRange(db *gorm.DB, from *time.Time, to *time.Time) ([]model.WorkUnitPeer, error)
}
