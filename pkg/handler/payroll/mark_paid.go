package payroll

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/store/payroll"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func (h *handler) MarkPayrollAsPaid(c *gin.Context) {
	var ids []string
	if err := c.Bind(&ids); err != nil {
		return
	}

	err := h.markPayrollAsPaid(ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
	}
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "marking the payroll as paid was successful"))
}

// markPayrollAsPaid from selected payroll row in database
// update the is_paid and is_sent_mail in payroll table
// send email into the users that marked as paid
func (h *handler) markPayrollAsPaid(ids []string) error {
	for _, id := range ids {
		q := payroll.GetListPayrollInput{
			ID: id,
		}
		ps, err := h.store.Payroll.GetList(h.repo.DB(), q)
		if err != nil {
			return err
		}
		if len(ps) == 0 {
			continue
		}

		if err := h.service.GoogleMail.SendPayrollPaidMail(&ps[0]); err != nil {
			return err
		}

		fields := map[string]interface{}{
			"is_paid": true,
		}
		if err := h.store.Payroll.UpdateSpecificFields(h.repo.DB(), id, fields); err != nil {
			return err
		}
	}
	return nil
}
