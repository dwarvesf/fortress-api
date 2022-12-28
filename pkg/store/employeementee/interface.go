package employeementee

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	OneByMenteeID(db *gorm.DB, id string, preload bool) (employeeMentee *model.EmployeeMentee, err error)
	Delete(db *gorm.DB, menteeID string) (err error)
	DeleteByMentorIDAndMenteeID(db *gorm.DB, mentorID string, menteeID string) (err error)
	Create(db *gorm.DB, e *model.EmployeeMentee) (employeeMentee *model.EmployeeMentee, err error)
	IsExist(db *gorm.DB, mentorID string, menteeID string) (exists bool, err error)
}
