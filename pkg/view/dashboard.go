package view

import (
	"sort"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type WorkUnitData struct {
	Development int64 `json:"development"`
	Management  int64 `json:"management"`
	Training    int64 `json:"training"`
	Learning    int64 `json:"learning"`
}

type WorkUnitDistributionData struct {
	Employee  BasicEmployeeInfo `json:"employee"`
	WorkUnits WorkUnitData      `json:"workunits"`
}

type WorkUnitDistributionDataList struct {
	WorkUnitDistributions []WorkUnitDistributionData `json:"workUnitDistributions"`
	Total                 WorkUnitData               `json:"total"`
}

type WorkUnitDistributionResponse struct {
	Data WorkUnitDistributionDataList `json:"data"`
}

func ToWorkUnitDistributionDataList(employees []*model.Employee, sortRequired model.SortOrder) WorkUnitDistributionDataList {
	rs := WorkUnitDistributionDataList{}
	total := WorkUnitData{}

	if sortRequired == model.SortOrderDESC {
		sort.Slice(employees, func(i, j int) bool {
			return len(employees[i].WorkUnitMembers) > len(employees[j].WorkUnitMembers)
		})
	} else if sortRequired == model.SortOrderASC {
		sort.Slice(employees, func(i, j int) bool {
			return len(employees[i].WorkUnitMembers) < len(employees[j].WorkUnitMembers)
		})
	}

	for _, e := range employees {
		workUnitDistributionData := WorkUnitDistributionData{}
		workUnitDistributionData.Employee = *toBasicEmployeeInfo(*e)

		for _, wu := range e.WorkUnitMembers {
			switch wu.WorkUnit.Type {
			case model.WorkUnitTypeDevelopment:
				total.Development++
				workUnitDistributionData.WorkUnits.Development++
			case model.WorkUnitTypeManagement:
				total.Management++
				workUnitDistributionData.WorkUnits.Management++
			case model.WorkUnitTypeTraining:
				total.Training++
				workUnitDistributionData.WorkUnits.Training++
			case model.WorkUnitTypeLearning:
				total.Learning++
				workUnitDistributionData.WorkUnits.Learning++
			}
		}

		rs.WorkUnitDistributions = append(rs.WorkUnitDistributions, workUnitDistributionData)
	}

	rs.Total = total
	return rs
}
