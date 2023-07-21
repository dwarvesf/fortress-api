package deliverymetrics

import (
	"github.com/pkg/errors"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (c controller) GetWeeklyLeaderBoard() (*model.WeeklyLeaderBoard, error) {
	w, err := c.store.DeliveryMetric.GetLatestWeek(c.repo.DB())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get latest week")
	}

	// Get top 10 users with highest points
	metrics, err := c.store.DeliveryMetric.GetTopWeighMetrics(c.repo.DB(), w, 5)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get top users with highest points")
	}

	items := make([]model.LeaderBoardItem, 0, len(metrics))
	// Get user info
	for _, m := range metrics {
		e, err := c.store.Employee.One(c.repo.DB(), m.EmployeeID.String(), false)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get employee "+m.EmployeeID.String())
		}

		// Get discord acc
		d, err := c.store.DiscordAccount.One(c.repo.DB(), e.DiscordAccountID.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get discord account "+e.DiscordAccountID.String()+" of employee "+e.ID.String())
		}

		items = append(items, model.LeaderBoardItem{
			EmployeeID:      e.ID.String(),
			EmployeeName:    e.DisplayName,
			Points:          m.Weight,
			Effectiveness:   m.Effectiveness,
			DiscordID:       d.DiscordID,
			DiscordUsername: d.Username,
		})
	}

	return &model.WeeklyLeaderBoard{
		Date:  w,
		Items: rankItems(items),
	}, nil
}

func rankItems(data []model.LeaderBoardItem) []model.LeaderBoardItem {
	// Set the rank for each employee
	for i := range data {
		if i > 0 && data[i].Points.Equal(data[i-1].Points) && data[i].Effectiveness.Equal(data[i-1].Effectiveness) {
			data[i].Rank = data[i-1].Rank
		} else {
			data[i].Rank = i + 1
		}
	}

	return data
}
