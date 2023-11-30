package payroll

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func (h *handler) GetPayrollsBHXH(c *gin.Context) {
	res, err := h.getPayrollBHXHHandler()
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(res, nil, nil, nil, ""))
}

func (h *handler) getPayrollBHXHHandler() (interface{}, error) {
	var res []payrollBHXHResponse

	isLeft := false
	for _, b := range []model.Batch{model.FirstBatch, model.SecondBatch} {
		date := time.Date(time.Now().Year(), time.Now().Month(), int(b), 0, 0, 0, 0, time.Now().Location())
		us, _, err := h.store.Employee.All(h.repo.DB(), employee.EmployeeFilter{IsLeft: &isLeft, BatchDate: &date, Preload: true}, model.Pagination{Page: 0, Size: 500})
		if err != nil {
			return nil, err
		}
		for i := range us {
			if us[i].BaseSalary.CompanyAccountAmount == 0 || us[i].BaseSalary.Batch != int(b) {
				continue
			}
			res = append(res, payrollBHXHResponse{
				DisplayName:   us[i].FullName,
				BHXH:          us[i].BaseSalary.CompanyAccountAmount,
				Batch:         int(b),
				AccountNumber: us[i].LocalBankNumber,
				Bank:          us[i].LocalBranchName,
			})
		}
	}

	return res, nil
}
