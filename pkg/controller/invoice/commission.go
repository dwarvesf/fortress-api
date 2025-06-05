package invoice

import (
	"fmt"
	"strings"
	"sync"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	bcConst "github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	storeemployeecommission "github.com/dwarvesf/fortress-api/pkg/store/employeecommission"
	storeinvoice "github.com/dwarvesf/fortress-api/pkg/store/invoice"
)

const (
	hiringCommissionRate       int64   = 2
	saleReferralCommissionRate int64   = 10
	inboundFundCommissionRate  float64 = 0.04 // 4%
)

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
	saleReferers     []pic
	upsellReferers   []pic
	dealClosing      []pic
}

func (c *controller) CalculateCommissionFromInvoice(db store.DBRepo, l logger.Logger, invoice *model.Invoice) ([]model.EmployeeCommission, error) {
	return c.calculateCommissionFromInvoice(db.DB(), l, invoice)
}

func (c *controller) storeCommission(db *gorm.DB, l logger.Logger, invoice *model.Invoice) ([]model.EmployeeCommission, error) {
	if invoice.Project.Type == model.ProjectTypeDwarves {
		return nil, nil
	}

	employeeCommissions, err := c.calculateCommissionFromInvoice(db, l, invoice)
	// employeeCommissions, err := c.calculateCommissionFromInvoice(db, l, invoice)
	if err != nil {
		l.Errorf(err, "failed to create commission for invoice(%s)", invoice.ID.String())
		return nil, err
	}

	if len(employeeCommissions) == 0 {
		return []model.EmployeeCommission{}, nil
	}

	for _, commission := range employeeCommissions {
		if strings.Contains(commission.Note, "Inbound Fund") {
			_, err := c.store.InboundFundTransaction.Create(db, &model.InboundFundTransaction{
				InvoiceID:      invoice.ID,
				Amount:         commission.Amount,
				ConversionRate: commission.ConversionRate,
				Notes:          fmt.Sprintf("%v - %v", commission.Formula, commission.Note),
			})
			if err != nil {
				l.Errorf(err, "failed to create inbound fund transaction for invoice(%s)", invoice.ID.String())
				return nil, err
			}
		}
	}
	// remove inbound fund commission from employee commissions
	employeeCommissions = c.RemoveInboundFundCommission(employeeCommissions)

	return c.store.EmployeeCommission.Create(db, employeeCommissions)
}

func (c *controller) RemoveInboundFundCommission(employeeCommissions []model.EmployeeCommission) []model.EmployeeCommission {
	rs := []model.EmployeeCommission{}
	for _, commission := range employeeCommissions {
		if !strings.Contains(commission.Note, "Inbound Fund") {
			rs = append(rs, commission)
		}
	}
	return rs
}

func (c *controller) calculateCommissionFromInvoice(db *gorm.DB, l logger.Logger, invoice *model.Invoice) ([]model.EmployeeCommission, error) {
	projectMembers, err := c.store.ProjectMember.GetAssignedMembers(db, invoice.ProjectID.String(), model.ProjectMemberStatusActive.String(), true)
	if err != nil {
		l.Errorf(err, "failed to calculate account manager commission rate for project(%s)", invoice.ProjectID.String())
		return nil, err
	}

	// Get list of project head who will get the commission from this invoice
	pics := c.getPICs(invoice, projectMembers)
	var res []model.EmployeeCommission
	if len(pics.devLeads) > 0 {
		c, err := c.calculateHeadCommission(pics.devLeads, invoice, invoice.TotalWithoutBonus)
		if err != nil {
			l.Errorf(err, "failed to calculate dev lead commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.accountManagers) > 0 {
		c, err := c.calculateHeadCommission(pics.accountManagers, invoice, invoice.TotalWithoutBonus)
		if err != nil {
			l.Errorf(err, "failed to calculate account manager commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.deliveryManagers) > 0 {
		c, err := c.calculateHeadCommission(pics.deliveryManagers, invoice, invoice.TotalWithoutBonus)
		if err != nil {
			l.Errorf(err, "failed to calculate delivery manager commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.sales) > 0 {
		c, err := c.calculateHeadCommission(pics.sales, invoice, invoice.TotalWithoutBonus)
		if err != nil {
			l.Errorf(err, "failed to calculate sales commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}

		res = append(res, c...)
	} else {
		// calculate commission for inbound
		c, err := c.calculateInboundFundCommission(invoice, invoice.TotalWithoutBonus)
		if err != nil {
			l.Errorf(err, "failed to calculate inbound fund commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.dealClosing) > 0 {
		c, err := c.calculateHeadCommission(pics.dealClosing, invoice, invoice.TotalWithoutBonus)
		if err != nil {
			l.Errorf(err, "failed to calculate deal closing commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.upsells) > 0 {
		c, err := c.calculateRefBonusCommission(pics.upsells, invoice)
		if err != nil {
			l.Errorf(err, "failed to calculate upsells commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.suppliers) > 0 {
		c, err := c.calculateRefBonusCommission(pics.suppliers, invoice)
		if err != nil {
			l.Errorf(err, "failed to calculate supplier commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.saleReferers) > 0 {
		c, err := c.calculateSaleReferralCommission(pics.saleReferers, invoice)
		if err != nil {
			l.Errorf(err, "failed to calculate sale refereral commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.upsellReferers) > 0 {
		c, err := c.calculateUpsellSaleReferralCommission(pics.upsellReferers, invoice)
		if err != nil {
			l.Errorf(err, "failed to calculate upsell refereral commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	return res, nil
}

func (c *controller) getPICs(invoice *model.Invoice, projectMembers []*model.ProjectMember) *pics {
	var (
		devLeads         []pic
		accountManagers  []pic
		deliveryManagers []pic
		sales            []pic
		upsells          []pic
		suppliers        []pic
		saleReferers     []pic
		upsellReferers   []pic
		dealClosing      []pic
	)

	for _, itm := range invoice.Project.Heads {
		switch itm.Position {
		case model.HeadPositionTechnicalLead:
			devLeadPersonDetail, err := c.store.Employee.One(c.repo.DB(), itm.EmployeeID.String(), false)
			if err != nil {
				continue
			}
			devLeads = append(devLeads, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     invoice.Total,
				Note:           fmt.Sprintf("Lead - %s", devLeadPersonDetail.FullName),
			})
		case model.HeadPositionAccountManager:
			accountManagers = append(accountManagers, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     invoice.Total,
				Note:           "Account Manager",
			})
		case model.HeadPositionDealClosing:
			dealClosingPersonDetail, err := c.store.Employee.One(c.repo.DB(), itm.EmployeeID.String(), false)
			if err != nil {
				continue
			}
			dealClosing = append(dealClosing, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     invoice.Total,
				Note:           fmt.Sprintf("Deal Closing - %s", dealClosingPersonDetail.FullName),
			})
		case model.HeadPositionDeliveryManager:
			deliveryManagers = append(deliveryManagers, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     invoice.Total,
				Note:           "Delivery Manager",
			})
		case model.HeadPositionSalePerson:
			// UPDATE:
			// If sale person earn commission from project. The person who refer the sale person will earn commission too.
			// The commission rate is 10% of the sale person commission rate.
			salePersonDetail, err := c.store.Employee.One(c.repo.DB(), itm.EmployeeID.String(), false)
			if err != nil {
				continue
			}

			sales = append(sales, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     invoice.Total,
				Note:           fmt.Sprintf("Sales - %s", salePersonDetail.FullName),
			})

			if salePersonDetail.Referrer == nil {
				continue
			}
			saleReferers = append(saleReferers, pic{
				ID:             salePersonDetail.Referrer.ID,
				CommissionRate: decimal.NewFromInt(saleReferralCommissionRate),
				ChargeRate:     decimal.NewFromFloat(invoice.Total).Mul(itm.CommissionRate).Div(decimal.NewFromInt(100)).InexactFloat64(),
				Note:           fmt.Sprintf("Sale Referral - %s", salePersonDetail.FullName),
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

			upsellPersonDetail, err := c.store.Employee.One(c.repo.DB(), pm.UpsellPersonID.String(), false)
			if err != nil {
				continue
			}

			if upsellPersonDetail.Referrer != nil {
				upsellReferers = append(upsellReferers, pic{
					ID:             upsellPersonDetail.Referrer.ID,
					CommissionRate: decimal.NewFromInt(saleReferralCommissionRate),
					ChargeRate:     decimal.NewFromFloat(pm.Rate.InexactFloat64()).Mul(pm.UpsellCommissionRate).Div(decimal.NewFromInt(100)).InexactFloat64(),
					Note:           fmt.Sprintf("Sale Referral - %s Upsell %s", upsellPersonDetail.FullName, pm.Employee.FullName),
				})
			}
		}

		if pm.Employee.Referrer != nil {
			if pm.Employee.Referrer.WorkingStatus != model.WorkingStatusLeft {
				suppliers = append(suppliers, pic{
					ID:             pm.Employee.Referrer.ID,
					CommissionRate: decimal.NewFromInt(hiringCommissionRate),
					ChargeRate:     pm.Rate.InexactFloat64(),
					Note:           fmt.Sprintf("Hiring - %s", pm.Employee.FullName),
				})
			}
		}
	}

	return &pics{
		devLeads:         devLeads,
		accountManagers:  accountManagers,
		deliveryManagers: deliveryManagers,
		sales:            sales,
		upsells:          upsells,
		suppliers:        suppliers,
		saleReferers:     saleReferers,
		upsellReferers:   upsellReferers,
		dealClosing:      dealClosing,
	}
}

func (c *controller) movePaidInvoiceGDrive(l logger.Logger, wg *sync.WaitGroup, req *processPaidInvoiceRequest) {
	msg := c.service.Basecamp.BuildCommentMessage(req.InvoiceBucketID, req.InvoiceTodoID, bcConst.CommentMoveInvoicePDFToPaidDirSuccessfully, bcModel.CommentMsgTypeCompleted)

	defer func() {
		c.worker.Enqueue(bcModel.BasecampCommentMsg, msg)
		wg.Done()
	}()

	err := c.service.GoogleDrive.MoveInvoicePDF(req.Invoice, "Sent", "Paid")
	if err != nil {
		l.Errorf(err, "failed to move invoice pdf from sent to paid folder for invoice(%v)", req.Invoice.ID.String())
		msg = c.service.Basecamp.BuildCommentMessage(req.InvoiceBucketID, req.InvoiceTodoID, bcConst.CommentMoveInvoicePDFToPaidDirFailed, bcModel.CommentMsgTypeFailed)
		return
	}
}

func (c *controller) calculateHeadCommission(beneficiaries []pic, invoice *model.Invoice, invoiceTotal float64) ([]model.EmployeeCommission, error) {
	// NOTE:
	// CR is Commission Rate
	// IV is Invoice Value
	// RCR is Referral Commission Rate
	// VAL is Charge Rate the value before the commission
	// RATE is Conversion Rate

	rs := make([]model.EmployeeCommission, 0)
	for _, beneficiary := range beneficiaries {
		if !beneficiary.CommissionRate.GreaterThan(decimal.NewFromInt(0)) {
			continue
		}

		crPercentage := beneficiary.CommissionRate.Div(decimal.NewFromInt(100))
		commissionValue, _ := crPercentage.Mul(decimal.NewFromFloat(invoiceTotal)).Float64()
		convertedValue, rate, err := c.service.Wise.Convert(commissionValue, invoice.Project.BankAccount.Currency.Name, "VND")
		if err != nil {
			return nil, err
		}

		if convertedValue > 0 {
			rs = append(rs, model.EmployeeCommission{
				EmployeeID:     beneficiary.ID,
				Amount:         model.NewVietnamDong(int64(convertedValue)),
				Project:        invoice.Project.Name,
				ConversionRate: rate,
				InvoiceID:      invoice.ID,
				Formula:        fmt.Sprintf("%v%%(CR) * %v(IV) * %v(RATE)", beneficiary.CommissionRate, invoiceTotal, rate),
				Note:           beneficiary.Note,
			})
		}
	}

	return rs, nil
}

func (c *controller) calculateInboundFundCommission(invoice *model.Invoice, invoiceTotal float64) ([]model.EmployeeCommission, error) {
	commissionValue, _ := decimal.NewFromFloat(inboundFundCommissionRate).Mul(decimal.NewFromFloat(invoiceTotal)).Float64()
	convertedValue, rate, err := c.service.Wise.Convert(commissionValue, invoice.Project.BankAccount.Currency.Name, "VND")
	if err != nil {
		return nil, err
	}

	rs := []model.EmployeeCommission{
		{
			EmployeeID:     model.UUID{},
			Amount:         model.NewVietnamDong(int64(convertedValue)),
			Project:        invoice.Project.Name,
			InvoiceID:      invoice.ID,
			ConversionRate: rate,
			Formula:        fmt.Sprintf("%v%%(CR) * %v(IV) * %v(RATE)", inboundFundCommissionRate*100, invoiceTotal, rate),
			Note:           fmt.Sprintf("Inbound Fund - %s", invoice.Number),
		},
	}

	return rs, nil
}

func (c *controller) calculateRefBonusCommission(pics []pic, invoice *model.Invoice) ([]model.EmployeeCommission, error) {
	// conversionRate by percentage
	var rs []model.EmployeeCommission
	for _, pic := range pics {
		percentage := pic.CommissionRate.Div(decimal.NewFromInt(100))
		commissionValue, _ := percentage.Mul(decimal.NewFromFloat(pic.ChargeRate)).Float64()
		convertedValue, rate, err := c.service.Wise.Convert(commissionValue, invoice.Project.BankAccount.Currency.Name, "VND")
		if err != nil {
			return nil, err
		}

		rs = append(rs, model.EmployeeCommission{
			EmployeeID:     pic.ID,
			Amount:         model.NewVietnamDong(int64(convertedValue)),
			Project:        invoice.Project.Name,
			ConversionRate: rate,
			InvoiceID:      invoice.ID,
			Formula:        fmt.Sprintf("%v%%(RCR) * %v(VAL) * %v(RATE)", pic.CommissionRate, pic.ChargeRate, rate),
			Note:           pic.Note,
		})
	}

	return rs, nil
}

func (c *controller) calculateSaleReferralCommission(pics []pic, invoice *model.Invoice) ([]model.EmployeeCommission, error) {
	// conversionRate by percentage
	var rs []model.EmployeeCommission
	for _, pic := range pics {
		percentage := pic.CommissionRate.Div(decimal.NewFromInt(100))
		commissionValue, _ := percentage.Mul(decimal.NewFromFloat(pic.ChargeRate)).Float64()
		convertedValue, rate, err := c.service.Wise.Convert(commissionValue, invoice.Project.BankAccount.Currency.Name, "VND")
		if err != nil {
			return nil, err
		}

		rs = append(rs, model.EmployeeCommission{
			EmployeeID:     pic.ID,
			Amount:         model.NewVietnamDong(int64(convertedValue)),
			Project:        invoice.Project.Name,
			ConversionRate: rate,
			InvoiceID:      invoice.ID,
			Formula:        fmt.Sprintf("%v%%(RCR) * %v(VAL) * %v(RATE)", pic.CommissionRate, commissionValue, rate),
			Note:           pic.Note,
		})
	}

	return rs, nil
}

func (c *controller) calculateUpsellSaleReferralCommission(pics []pic, invoice *model.Invoice) ([]model.EmployeeCommission, error) {
	var rs []model.EmployeeCommission
	for _, pic := range pics {
		percentage := pic.CommissionRate.Div(decimal.NewFromInt(100))
		commissionValue, _ := percentage.Mul(decimal.NewFromFloat(pic.ChargeRate)).Float64()
		convertedValue, rate, err := c.service.Wise.Convert(commissionValue, invoice.Project.BankAccount.Currency.Name, "VND")
		if err != nil {
			return nil, err
		}

		rs = append(rs, model.EmployeeCommission{
			EmployeeID:     pic.ID,
			Amount:         model.NewVietnamDong(int64(convertedValue)),
			Project:        invoice.Project.Name,
			ConversionRate: rate,
			InvoiceID:      invoice.ID,
			Formula:        fmt.Sprintf("%v%%(RCR) * %v(VAL) * %v(RATE)", pic.CommissionRate, commissionValue, rate),
			Note:           pic.Note,
		})
	}

	return rs, nil
}

// ProcessCommissions calculates and (optionally) saves commissions for an invoice in a transaction-safe way.
func (c *controller) ProcessCommissions(invoiceID string, dryRun bool, l logger.Logger) ([]model.EmployeeCommission, error) {
	db := c.repo.DB()
	tx := db.Begin()
	if tx.Error != nil {
		l.Error(tx.Error, "failed to begin database transaction")
		return nil, tx.Error
	}

	// Fetch invoice with all required preloads
	invoice, err := c.store.Invoice.One(tx, &storeinvoice.Query{ID: invoiceID})
	if err != nil {
		tx.Rollback()
		l.Error(err, "failed to find invoice in database")
		return nil, err
	}

	commissions, err := c.CalculateCommissionFromInvoice(c.repo, l, invoice)
	if err != nil {
		tx.Rollback()
		l.Error(err, "failed to calculate commissions")
		return nil, err
	}

	if dryRun {
		tx.Rollback()
		return commissions, nil
	}

	// Only delete and create if there are unpaid records
	if err := c.store.EmployeeCommission.DeleteUnpaidByInvoiceID(tx, invoiceID); err != nil {
		tx.Rollback()
		l.Error(err, "failed to delete existing commissions")
		return nil, err
	}
	if err := c.store.InboundFundTransaction.DeleteUnpaidByInvoiceID(tx, invoiceID); err != nil {
		tx.Rollback()
		l.Error(err, "failed to delete existing inbound fund transactions")
		return nil, err
	}

	// Create inbound fund transactions
	for _, commission := range commissions {
		if commission.EmployeeID.IsZero() {
			inboundFunCommission, err := c.store.InboundFundTransaction.GetByInvoiceID(tx, invoiceID)
			if err == nil && inboundFunCommission.PaidAt != nil {
				continue // already paid, skip creation
			}
			_, err = c.store.InboundFundTransaction.Create(tx, &model.InboundFundTransaction{
				InvoiceID:      invoice.ID,
				Amount:         commission.Amount,
				ConversionRate: commission.ConversionRate,
				Notes:          fmt.Sprintf("%v - %v", commission.Formula, commission.Note),
			})
			if err != nil {
				tx.Rollback()
				l.Errorf(err, "failed to create inbound fund transaction for invoice(%s)", invoice.ID.String())
				return nil, err
			}
		}
	}

	// Remove inbound fund commissions from employee commissions
	filtered := c.RemoveInboundFundCommission(commissions)
	if len(filtered) > 0 {
		for _, commission := range filtered {
			isPaidCommission, err := c.store.EmployeeCommission.Get(tx, storeemployeecommission.Query{
				InvoiceID:  commission.InvoiceID.String(),
				EmployeeID: commission.EmployeeID.String(),
				IsPaid:     true,
			})
			if err != nil {
				tx.Rollback()
				l.Error(err, "failed to check if commission is already paid")
				continue
			}
			if len(isPaidCommission) > 0 {
				l.Infof("Commission for invoice %s and employee %s is already paid, skipping creation", commission.InvoiceID, commission.EmployeeID)
				continue // already paid, skip creation
			}
			if _, err := c.store.EmployeeCommission.Create(tx, []model.EmployeeCommission{commission}); err != nil {
				tx.Rollback()
				l.Error(err, "failed to create commission record")
				return nil, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		l.Error(err, "failed to commit transaction")
		return nil, err
	}

	return filtered, nil
}
