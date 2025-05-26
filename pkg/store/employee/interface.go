package employee

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB, input EmployeeFilter, pagination model.Pagination) (employees []*model.Employee, total int64, err error)
	Create(db *gorm.DB, e *model.Employee) (employee *model.Employee, err error)

	One(db *gorm.DB, id string, preload bool) (employee *model.Employee, err error)
	OneByEmail(db *gorm.DB, email string) (*model.Employee, error)
	OneByNotionID(db *gorm.DB, notionID string) (employee *model.Employee, err error)
	OneByBasecampID(db *gorm.DB, basecampID int) (employee *model.Employee, err error)
	GetByIDs(db *gorm.DB, ids []model.UUID) (employees []*model.Employee, err error)
	GetByEmails(db *gorm.DB, emails []string) (employees []*model.Employee, err error)
	GetByBasecampIDs(db *gorm.DB, basecampIDs []int) (employees []*model.Employee, err error)
	GetByWorkingStatus(db *gorm.DB, workingStatus model.WorkingStatus) ([]*model.Employee, error)
	GetLineManagers(db *gorm.DB) ([]*model.Employee, error)
	GetLineManagersOfPeers(db *gorm.DB, employeeID string) ([]*model.Employee, error)
	GetMenteesByID(db *gorm.DB, employeeID string) ([]*model.Employee, error)
	GetByDiscordID(db *gorm.DB, discordID string, preload bool) (*model.Employee, error)
	GetByDiscordUsername(db *gorm.DB, discordUsername string) (*model.Employee, error)
	ListByDiscordRequest(db *gorm.DB, in DiscordRequestFilter, preload bool) ([]model.Employee, error)
	ListWithMMAScore(db *gorm.DB) ([]model.EmployeeMMAScoreData, error)
	SimpleList(db *gorm.DB) ([]*model.Employee, error)
	GetRawList(db *gorm.DB, filter EmployeeFilter) ([]model.Employee, error)

	IsExist(db *gorm.DB, id string) (bool, error)

	Update(db *gorm.DB, employee *model.Employee) (*model.Employee, error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Employee, updatedFields ...string) (*model.Employee, error)

	OneByDisplayName(db *gorm.DB, displayName string) (*model.Employee, error)
}
