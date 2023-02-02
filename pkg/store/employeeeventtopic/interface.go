package employeeeventtopic

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string, eventID string, preload bool) (topic *model.EmployeeEventTopic, err error)
	All(db *gorm.DB, input GetByEventIDInput, pagination *model.Pagination) (topics []*model.EmployeeEventTopic, total int64, err error)
	GetByEmployeeID(db *gorm.DB, employeeID string, input GetByEmployeeIDInput, pagination model.Pagination) (eTopics []*model.EmployeeEventTopic, total int64, err error)
	BatchCreate(db *gorm.DB, employeeEventTopics []model.EmployeeEventTopic) ([]model.EmployeeEventTopic, error)
	DeleteByEventID(db *gorm.DB, eventID string) error
	DeleteByID(db *gorm.DB, ID string) error
	CountUnreadFeedbackByEmployeeID(db *gorm.DB, employeeID string) (count int64, err error)
}
