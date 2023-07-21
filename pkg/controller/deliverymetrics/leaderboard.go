package deliverymetrics

import (
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type WeeklyLeaderBoard struct {
	Date  *time.Time        `json:"date"`
	Items []LeaderBoardItem `json:"items"`
}

type LeaderBoardItem struct {
	EmployeeID   string          `json:"employee_id"`
	EmployeeName string          `json:"employee_name"`
	Points       decimal.Decimal `json:"points"`

	DiscordID       string `json:"discord_id"`
	DiscordUsername string `json:"discord_username"`
}

func (c controller) GetWeeklyLeaderBoard() (*WeeklyLeaderBoard, error) {
	w, err := c.store.DeliveryMetric.GetLatestWeek(c.repo.DB())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get latest week")
	}

	// Get top 10 users with highest points
	metrics, err := c.store.DeliveryMetric.GetTopWeighMetrics(c.repo.DB(), w, 10)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get top users with highest points")
	}

	items := make([]LeaderBoardItem, 0, len(metrics))
	// Get user info
	for _, m := range metrics {
		e, err := c.store.Employee.One(c.repo.DB(), m.EmployeeID.String(), true)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get employee "+m.EmployeeID.String())
		}

		// Get discord acc
		d, err := c.store.DiscordAccount.One(c.repo.DB(), e.DiscordAccountID.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get discord account "+e.DiscordAccountID.String()+" of employee "+e.ID.String())
		}

		items = append(items, LeaderBoardItem{
			EmployeeID:      e.ID.String(),
			EmployeeName:    e.DisplayName,
			Points:          m.Weight,
			DiscordID:       d.DiscordID,
			DiscordUsername: d.Username,
		})
	}

	return &WeeklyLeaderBoard{
		Date:  w,
		Items: items,
	}, nil
}
