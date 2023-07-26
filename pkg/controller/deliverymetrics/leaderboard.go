package deliverymetrics

import (
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (c controller) GetWeeklyLeaderBoard() (*model.LeaderBoard, error) {
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

		item := model.LeaderBoardItem{
			EmployeeID:      e.ID.String(),
			EmployeeName:    e.DisplayName,
			Points:          m.SumWeight,
			DiscordID:       d.DiscordID,
			DiscordUsername: d.Username,
		}
		if m.SumEffort.IsZero() {
			item.Effectiveness = decimal.NewFromFloat(0)
		} else {
			item.Effectiveness = m.SumWeight.DivRound(m.SumEffort, 2)
		}

		items = append(items, item)
	}

	return &model.LeaderBoard{
		Date:  w,
		Items: rankItems(items),
	}, nil
}

func (c controller) GetMonthlyLeaderBoard(month *time.Time) (*model.LeaderBoard, error) {
	m := month
	if m == nil {
		var err error
		m, err = c.store.DeliveryMetric.GetLatestMonth(c.repo.DB())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get latest month")
		}
	}

	// Get top 10 users with highest points
	metrics, err := c.store.DeliveryMetric.GetTopMonthlyWeighMetrics(c.repo.DB(), m, 10)
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

		item := model.LeaderBoardItem{
			EmployeeID:    e.ID.String(),
			EmployeeName:  e.DisplayName,
			Points:        m.Weight,
			Effectiveness: m.Effectiveness,
		}

		// Get discord acc
		if !e.DiscordAccountID.IsZero() {
			d, err := c.store.DiscordAccount.One(c.repo.DB(), e.DiscordAccountID.String())
			if err != nil {
				return nil, errors.Wrap(err, "failed to get discord account "+e.DiscordAccountID.String()+" of employee "+e.ID.String())
			}
			if d != nil {
				item.DiscordID = d.DiscordID
				item.DiscordUsername = d.Username
			}
		}

		items = append(items, item)
	}

	return &model.LeaderBoard{
		Date:  m,
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
