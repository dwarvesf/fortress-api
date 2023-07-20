package deliverymetrics

import (
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
)

func (c controller) Sync() error {
	l := c.logger.Fields(logger.Fields{
		"controller": "deliverymetrics",
		"method":     "Create",
	})

	latestItem, err := c.store.DeliveryMetric.GetLatest(c.repo.DB())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorf(err, "failed to get latest item")
		return err
	}

	maxIdx := 0
	if errors.Is(err, gorm.ErrRecordNotFound) {
		maxIdx = 2
	} else {
		maxIdx = latestItem.Ref + 1
	}

	sheetData, err := c.service.GoogleSheet.FetchSheetContent(maxIdx)
	if err != nil {
		l.Errorf(err, "failed to fetch sheet content")
		return err
	}

	employees, err := c.store.Employee.GetRawList(c.repo.DB(), employee.EmployeeFilter{})
	if err != nil {
		l.Errorf(err, "failed to fetch employee")
		return err
	}
	employeeMap := model.Employees(employees).ToTeamEmailIDMap()

	projects, err := c.store.Project.GetRawList(c.repo.DB())
	if err != nil {
		l.Errorf(err, "failed to fetch project")
		return err
	}

	projectMap := model.Projects(projects).ToNameIDMap()

	deliveryMetrics := make([]model.DeliveryMetric, 0)
	for _, row := range sheetData {
		defaultVal := decimal.NewFromInt(0)
		weight, err := decimal.NewFromString(strings.ReplaceAll(row.Weight, ",", "."))
		if err != nil {
			weight = defaultVal
		}

		effort, err := decimal.NewFromString(strings.ReplaceAll(row.Effort, ",", "."))
		if err != nil {
			effort = defaultVal
		}

		effectiveness, err := decimal.NewFromString(strings.ReplaceAll(row.Effectiveness, ",", "."))
		if err != nil {
			effectiveness = defaultVal
		}

		date, err := time.Parse("01-02-2006", row.Date)
		if err != nil {
			continue
		}

		projectName := row.Project
		if row.Project == "Internal" {
			projectName = "Fortress"
		}

		projectID := projectMap[projectName]
		employeeID := employeeMap[row.Email]

		dm := model.DeliveryMetric{
			Weight:        weight,
			Effort:        effort,
			Effectiveness: effectiveness,
			EmployeeID:    employeeID,
			ProjectID:     projectID,
			Date:          &date,
			Ref:           maxIdx,
		}

		deliveryMetrics = append(deliveryMetrics, dm)
		maxIdx++
	}

	if len(deliveryMetrics) > 0 {
		_, err = c.store.DeliveryMetric.Create(c.repo.DB(), deliveryMetrics)
		if err != nil {
			l.Errorf(err, "failed to create delivery metric")
			return err
		}
	}

	return nil
}
