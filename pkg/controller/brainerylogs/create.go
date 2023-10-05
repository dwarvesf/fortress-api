package brainerylogs

import (
	"errors"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

// Create creates a new brainery log
func (c *controller) Create(log model.BraineryLog) (model.BraineryLog, error) {
	emp, err := c.store.Employee.GetByDiscordID(c.repo.DB(), log.DiscordID, false)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.logger.Errorf(err, "failed to get employee by discordID", "discordID", log.DiscordID)
		return model.BraineryLog{}, err
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.EmployeeID = emp.ID
	}

	_, err = c.store.BraineryLog.Create(c.repo.DB(), []model.BraineryLog{log})
	if err != nil {
		c.logger.Errorf(err, "failed to create brainery logs", "braineryLog", log)
		return model.BraineryLog{}, err
	}

	return log, nil
}
