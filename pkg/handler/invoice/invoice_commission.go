package invoice

import (
	"fmt"
	"sync"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	bcConst "github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

const hiringCommissionRate int64 = 2

type pic struct {
	ID             model.UUID
	CommissionRate decimal.Decimal
	ChargeRate     float64
	Note           string
}

type pics struct {
	devLeads         []pic
	accountManagers  []pic
	deliveryManagers []pic
	sales            []pic
	suppliers        []pic
	upsells          []pic
}

func (h *handler) storeCommission(db *gorm.DB, l logger.Logger, invoice *model.Invoice) ([]model.EmployeeCommission, error) {
	if invoice.Project.Type != model.ProjectTypeTimeMaterial {
		return nil, nil
	}
	employeeCommissions, err := h.calculateCommissionFromInvoice(db, l, invoice)
	if err != nil {
		l.Errorf(err, "failed to create commission", "invoice", invoice)
		return nil, err
	}

	return h.store.EmployeeCommission.Create(db, employeeCommissions)
}

func (h *handler) calculateCommissionFromInvoice(db *gorm.DB, l logger.Logger, invoice *model.Invoice) ([]model.EmployeeCommission, error) {
	// Get project commission configs
	commissionConfigs, err := h.store.ProjectCommissionConfig.GetByProjectID(db, invoice.ProjectID.String())
	if err != nil {
		l.Errorf(err, "failed to get project commission config", "projectID", invoice.ProjectID.String())
		return nil, err
	}
	commissionConfigMap := commissionConfigs.ToMap()

	projectMembers, err := h.store.ProjectMember.GetAssignedMembers(db, invoice.ProjectID.String(), model.ProjectMemberStatusActive.String(), true)
	if err != nil {
		l.Errorf(err, "failed to calculate account manager commission rate", "projectID", invoice.ProjectID.String())
		return nil, err
	}

	// Get list of project head who will get the commission from this invoice
	pics := getPICs(invoice, projectMembers)

	var res []model.EmployeeCommission
	if len(pics.devLeads) > 0 {
		commissionRate := commissionConfigMap[model.HeadPositionTechnicalLead.String()]
		if commissionRate.GreaterThan(decimal.NewFromInt(0)) {
			c, err := h.calculateHeadCommission(commissionRate, pics.devLeads, invoice, float64(invoice.Total))
			if err != nil {
				l.Errorf(err, "failed to calculate dev lead commission rate", "projectID", invoice.ProjectID.String())
				return nil, err
			}
			res = append(res, c...)
		}
	}

	if len(pics.accountManagers) > 0 {
		commissionRate := commissionConfigMap[model.HeadPositionAccountManager.String()]
		if commissionRate.GreaterThan(decimal.NewFromInt(0)) {
			c, err := h.calculateHeadCommission(commissionRate, pics.accountManagers, invoice, float64(invoice.Total))
			if err != nil {
				l.Errorf(err, "failed to calculate account manager commission rate", "projectID", invoice.ProjectID.String())
				return nil, err
			}
			res = append(res, c...)
		}
	}

	if len(pics.deliveryManagers) > 0 {
		commissionRate := commissionConfigMap[model.HeadPositionDeliveryManager.String()]
		if commissionRate.GreaterThan(decimal.NewFromInt(0)) {
			c, err := h.calculateHeadCommission(commissionRate, pics.deliveryManagers, invoice, float64(invoice.Total))
			if err != nil {
				l.Errorf(err, "failed to calculate delivery manager commission rate", "projectID", invoice.ProjectID.String())
				return nil, err
			}
			res = append(res, c...)
		}
	}

	if len(pics.sales) > 0 {
		commissionRate := commissionConfigMap[model.HeadPositionSalePerson.String()]
		if commissionRate.GreaterThan(decimal.NewFromInt(0)) {
			c, err := h.calculateHeadCommission(commissionRate, pics.sales, invoice, float64(invoice.Total))
			if err != nil {
				l.Errorf(err, "failed to calculate account manager commission rate", "projectID", invoice.ProjectID.String())
				return nil, err
			}
			res = append(res, c...)
		}
	}

	if len(pics.upsells) > 0 {
		c, err := h.calculateRefBonusCommission(pics.upsells, invoice)
		if err != nil {
			l.Errorf(err, "failed to calculate account manager commission rate", "projectID", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.suppliers) > 0 {
		c, err := h.calculateRefBonusCommission(pics.suppliers, invoice)
		if err != nil {
			l.Errorf(err, "failed to calculate account manager commission rate", "projectID", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	return res, nil
}

func getPICs(invoice *model.Invoice, projectMembers []*model.ProjectMember) *pics {
	var (
		devLeads         []pic
		accountManagers  []pic
		deliveryManagers []pic
		sales            []pic
		upsells          []pic
		suppliers        []pic
	)

	for _, itm := range invoice.Project.Heads {
		switch itm.Position {
		case model.HeadPositionTechnicalLead:
			devLeads = append(devLeads, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     float64(invoice.Total),
				Note:           "Lead",
			})
		case model.HeadPositionAccountManager:
			accountManagers = append(accountManagers, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     float64(invoice.Total),
				Note:           "Account Manager",
			})
		case model.HeadPositionDeliveryManager:
			deliveryManagers = append(deliveryManagers, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     float64(invoice.Total),
				Note:           "Delivery Manager",
			})
		case model.HeadPositionSalePerson:
			sales = append(sales, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     float64(invoice.Total),
				Note:           "Sales",
			})
		}
	}

	for _, pm := range projectMembers {
		if pm.DeploymentType != model.MemberDeploymentTypeOfficial {
			continue
		}

		if !pm.UpsellCommissionRate.IsZero() {
			upsells = append(upsells, pic{
				ID:             pm.UpsellPersonID,
				CommissionRate: pm.UpsellCommissionRate,
				ChargeRate:     pm.Rate.InexactFloat64(),
				Note:           "Upsell",
			})
		}

		if pm.Employee.Referrer != nil {
			suppliers = append(suppliers, pic{
				ID:             pm.Employee.Referrer.ID,
				CommissionRate: decimal.NewFromInt(hiringCommissionRate),
				ChargeRate:     pm.Rate.InexactFloat64(),
				Note:           pm.Employee.FullName,
			})
		}
	}

	return &pics{
		devLeads:         devLeads,
		accountManagers:  accountManagers,
		deliveryManagers: deliveryManagers,
		sales:            sales,
		upsells:          upsells,
		suppliers:        suppliers,
	}
}

func (h *handler) movePaidInvoiceGDrive(l logger.Logger, wg *sync.WaitGroup, req *processPaidInvoiceRequest) {
	msg := h.service.Basecamp.BuildCommentMessage(req.InvoiceBucketID, req.InvoiceTodoID, bcConst.CommentMoveInvoicePDFToPaidDirSuccessfully, bcModel.CommentMsgTypeCompleted)

	defer func() {
		h.worker.Enqueue(bcModel.BasecampCommentMsg, msg)
		wg.Done()
	}()

	err := h.service.GoogleDrive.MoveInvoicePDF(req.Invoice, "Sent", "Paid")
	if err != nil {
		l.Errorf(err, "failed to move invoice pdf from sent to paid folder", "invoice", req.Invoice)
		msg = h.service.Basecamp.BuildCommentMessage(req.InvoiceBucketID, req.InvoiceTodoID, bcConst.CommentMoveInvoicePDFToPaidDirFailed, bcModel.CommentMsgTypeFailed)
		return
	}
}

func (h *handler) calculateHeadCommission(projectCommissionRate decimal.Decimal, beneficiaries []pic, invoice *model.Invoice, invoiceTotalPrice float64) ([]model.EmployeeCommission, error) {
	// conversionRate by percentage
	pcrPercentage := projectCommissionRate.Div(decimal.NewFromInt(100))
	projectCommissionValue, _ := pcrPercentage.Mul(decimal.NewFromFloat(invoiceTotalPrice)).Float64()
	convertedValue, rate, err := h.service.Wise.Convert(projectCommissionValue, invoice.Project.BankAccount.Currency.Name, "VND")
	if err != nil {
		return nil, err
	}

	rs := make([]model.EmployeeCommission, 0)
	for _, beneficiary := range beneficiaries {
		picPercentage := beneficiary.CommissionRate.Div(decimal.NewFromInt(100))
		picCommissionValue, _ := picPercentage.Mul(decimal.NewFromFloat(convertedValue)).Float64()

		if picCommissionValue > 0 {
			rs = append(rs, model.EmployeeCommission{
				EmployeeID:     beneficiary.ID,
				Amount:         model.NewVietnamDong(int64(picCommissionValue)),
				Project:        invoice.Project.Name,
				ConversionRate: rate,
				InvoiceID:      invoice.ID,
				Formula:        fmt.Sprintf("%v%%(PCR) * %v%%(SCR) * %v(IV) * %v(RATE)", projectCommissionRate, beneficiary.CommissionRate, invoiceTotalPrice, rate),
				Note:           beneficiary.Note,
			})
		}
	}

	return rs, nil
}

func (h *handler) calculateRefBonusCommission(pics []pic, invoice *model.Invoice) ([]model.EmployeeCommission, error) {
	// conversionRate by percentage
	var rs []model.EmployeeCommission
	for _, pic := range pics {
		percentage := pic.CommissionRate.Div(decimal.NewFromInt(100))
		commissionValue, _ := percentage.Mul(decimal.NewFromFloat(pic.ChargeRate)).Float64()
		convertedValue, rate, err := h.service.Wise.Convert(commissionValue, invoice.Project.BankAccount.Currency.Name, "VND")
		if err != nil {
			return nil, err
		}

		rs = append(rs, model.EmployeeCommission{
			EmployeeID:     pic.ID,
			Amount:         model.NewVietnamDong(int64(convertedValue)),
			Project:        invoice.Project.Name,
			ConversionRate: rate,
			InvoiceID:      invoice.ID,
			Formula:        fmt.Sprintf("%v%%(RCR) * %v(CR) * %v(RATE)", pic.CommissionRate, pic.ChargeRate, rate),
			Note:           pic.Note,
		})
	}

	return rs, nil
}
