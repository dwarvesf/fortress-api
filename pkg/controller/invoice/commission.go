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

const (
	hiringCommissionRate       int64   = 2
	saleReferralCommissionRate int64   = 10
	inboundFundCommissionRate  float64 = 0.05 // 5%
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
}

func (c *controller) storeCommission(db *gorm.DB, l logger.Logger, invoice *model.Invoice) ([]model.EmployeeCommission, error) {
	if invoice.Project.Type != model.ProjectTypeTimeMaterial {
		return nil, nil
	}

	// Check for project head salesperson and apply inbound fund logic
	foundSalesPerson := false
	for _, head := range invoice.Project.Heads {
		if head.Position == model.HeadPositionSalePerson {
			foundSalesPerson = true
			break
		}
	}

	originalInvoiceTotal := invoice.Total // Store original total before modification
	if !foundSalesPerson {
		inboundAmount := originalInvoiceTotal * inboundFundCommissionRate
		invoice.InboundFundAmount = inboundAmount

		ift := &model.InboundFundTransaction{
			InvoiceID: invoice.ID,
			Amount:    inboundAmount,
			Notes:     fmt.Sprintf("5%% of invoice %s (%.2f) for project %s (ID: %s) with no sales head", invoice.Number, originalInvoiceTotal, invoice.Project.Name, invoice.ProjectID.String()),
		}
		_, err := c.store.InboundFundTransaction.Create(db, ift)
		if err != nil {
			l.Errorf(err, "failed to create inbound fund transaction for invoice(%s)", invoice.ID.String())
			return nil, err
		}

		// Update the invoice in DB with InboundFundAmount and adjusted TotalWithoutBonus
		updatedFields := []string{"InboundFundAmount"}
		// If InboundFundAmount was already non-zero, this might indicate a re-processing.
		// For simplicity, we're overwriting. Consider adding a check if this isn't desired.
		if _, err := c.store.Invoice.UpdateSelectedFieldsByID(db, invoice.ID.String(), *invoice, updatedFields...); err != nil {
			l.Errorf(err, "failed to update invoice with inbound fund details for invoice ID: %s", invoice.ID.String())
			return nil, err
		}
	}

	employeeCommissions, err := c.calculateCommissionFromInvoice(db, l, invoice)
	if err != nil {
		l.Errorf(err, "failed to create commission for invoice(%s)", invoice.ID.String())
		return nil, err
	}

	if len(employeeCommissions) == 0 {
		return []model.EmployeeCommission{}, nil
	}

	return c.store.EmployeeCommission.Create(db, employeeCommissions)
}

func (c *controller) calculateCommissionFromInvoice(db *gorm.DB, l logger.Logger, invoice *model.Invoice) ([]model.EmployeeCommission, error) {
	projectMembers, err := c.store.ProjectMember.GetAssignedMembers(db, invoice.ProjectID.String(), model.ProjectMemberStatusActive.String(), true)
	if err != nil {
		l.Errorf(err, "failed to calculate account manager commission rate for project(%s)", invoice.ProjectID.String())
		return nil, err
	}

	invoiceWithoutInbound := invoice.TotalWithoutBonus - invoice.InboundFundAmount

	// Get list of project head who will get the commission from this invoice
	pics := c.getPICs(invoice, projectMembers)
	var res []model.EmployeeCommission
	if len(pics.devLeads) > 0 {
		c, err := c.calculateHeadCommission(pics.devLeads, invoice, invoiceWithoutInbound)
		if err != nil {
			l.Errorf(err, "failed to calculate dev lead commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.accountManagers) > 0 {
		c, err := c.calculateHeadCommission(pics.accountManagers, invoice, invoiceWithoutInbound)
		if err != nil {
			l.Errorf(err, "failed to calculate account manager commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.deliveryManagers) > 0 {
		c, err := c.calculateHeadCommission(pics.deliveryManagers, invoice, invoiceWithoutInbound)
		if err != nil {
			l.Errorf(err, "failed to calculate delivery manager commission rate for project(%s)", invoice.ProjectID.String())
			return nil, err
		}
		res = append(res, c...)
	}

	if len(pics.sales) > 0 {
		c, err := c.calculateHeadCommission(pics.sales, invoice, invoiceWithoutInbound)
		if err != nil {
			l.Errorf(err, "failed to calculate sales commission rate for project(%s)", invoice.ProjectID.String())
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
	)

	for _, itm := range invoice.Project.Heads {
		switch itm.Position {
		case model.HeadPositionTechnicalLead:
			devLeads = append(devLeads, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     invoice.Total,
				Note:           "Lead",
			})
		case model.HeadPositionAccountManager:
			accountManagers = append(accountManagers, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     invoice.Total,
				Note:           "Account Manager",
			})
		case model.HeadPositionDeliveryManager:
			deliveryManagers = append(deliveryManagers, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     invoice.Total,
				Note:           "Delivery Manager",
			})
		case model.HeadPositionSalePerson:
			sales = append(sales, pic{
				ID:             itm.EmployeeID,
				CommissionRate: itm.CommissionRate,
				ChargeRate:     invoice.Total,
				Note:           "Sales",
			})

			// UPDATE:
			// If sale person earn commission from project. The person who refer the sale person will earn commission too.
			// The commission rate is 10% of the sale person commission rate.
			salePersonDetail, err := c.store.Employee.One(c.repo.DB(), itm.EmployeeID.String(), false)
			if err != nil {
				continue
			}

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
