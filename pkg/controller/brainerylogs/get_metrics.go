package brainerylogs

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

// GetMetrics returns brainery metrics
func (c *controller) GetMetrics(selectedDate time.Time, queryView string) (latestPosts []*model.BraineryLog, logs []*model.BraineryLog, ncids []string, err error) {
	l := c.logger.Fields(logger.Fields{
		"controller": "brainerylogs",
		"method":     "GetBraineryMetrics",
	})

	end := timeutil.GetEndDayOfWeek(selectedDate)
	start := timeutil.GetStartDayOfWeek(selectedDate)
	if queryView == "monthly" {
		start = timeutil.FirstDayOfMonth(int(selectedDate.Month()), selectedDate.Year())
		end = timeutil.LastDayOfMonth(int(selectedDate.Month()), selectedDate.Year())
	}

	// latest 10 posts
	latestPosts, err = c.store.BraineryLog.GetLimitByTimeRange(c.repo.DB(), &time.Time{}, &end, 10)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorf(err, "failed to get latest posts by time range", "start", start, "end", end)
		return nil, nil, nil, err
	}

	// weekly or monthly posts
	logs, err = c.store.BraineryLog.GetLimitByTimeRange(c.repo.DB(), &start, &end, 1000)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorf(err, "failed to get logs by time range", "start", start, "end", end)
		return nil, nil, nil, err
	}

	// ncids = new contributor discord IDs
	ncids, err = c.store.BraineryLog.GetNewContributorDiscordIDs(c.repo.DB(), &start, &end)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorf(err, "failed to get new contributor discord IDs by time range", "start", start, "end", end)
		return nil, nil, nil, err
	}

	return latestPosts, logs, ncids, nil
}
