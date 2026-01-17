package invoice

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Rhymond/go-money"
	toPdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

// ContractorInvoiceData represents the data for generating a contractor invoice
type ContractorInvoiceData struct {
	InvoiceNumber     string
	ContractorName    string
	Month             string
	Date              time.Time
	DueDate           time.Time
	Description       string // Description under BILL TO section
	BillingType       string
	Currency          string
	LineItems         []ContractorInvoiceLineItem
	MonthlyFixed      float64
	MonthlyFixedUSD   float64 // Monthly fixed amount converted to USD
	Total             float64
	TotalUSD          float64 // Total converted to USD
	ExchangeRate      float64 // Exchange rate used for conversion
	BankAccountHolder string
	BankName          string
	BankAccountNumber string
	BankSwiftBIC      string
	BankBranch        string

	// Multi-currency subtotal fields (added for multi-currency support)
	SubtotalVND        float64 // Sum of all VND-denominated items
	SubtotalUSDFromVND float64 // SubtotalVND converted to USD at ExchangeRate
	SubtotalUSDItems   float64 // Sum of all USD-denominated items
	SubtotalUSD        float64 // SubtotalUSDFromVND + SubtotalUSDItems

	// Notion relation IDs (for creating Contractor Payables record)
	ContractorPageID string   // Contractor page ID from rates query
	PayoutPageIDs    []string // Payout Item page IDs from pending payouts

	// PayDay for Period calculation
	PayDay int // Pay day of month (1 or 15)
}

// ContractorInvoiceLineItem represents a line item in a contractor invoice
type ContractorInvoiceLineItem struct {
	Title       string
	Description string  // Proof of Work
	Hours       float64 // Only for Hourly Rate
	Rate        float64 // Only for Hourly Rate
	Amount      float64 // Only for Hourly Rate
	AmountUSD   float64 // Amount converted to USD (Only for Hourly Rate)
	Type        string  // Payout source type (Contractor Payroll, Commission, Refund, etc.)

	// Commission-specific fields (for grouping)
	CommissionRole    string
	CommissionProject string

	// Original currency fields (added for multi-currency support)
	OriginalAmount   float64 // Amount in original currency (VND or USD)
	OriginalCurrency string  // "VND" or "USD"

	// Hourly rate metadata
	IsHourlyRate  bool   // Mark as hourly-rate Service Fee (for aggregation)
	ServiceRateID string // Contractor Rate page ID (for logging/debugging)
	TaskOrderID   string // Task Order Log page ID (for logging/debugging)
}

// ContractorInvoiceSection represents a section of line items in the invoice
type ContractorInvoiceSection struct {
	Name         string                      // "Development work from [start] to [end]", "Refund", "Bonus"
	IsAggregated bool                        // true for Development Work (Service Fee)
	Total        float64                     // Total amount for aggregated sections
	Currency     string                      // Currency for aggregated sections
	Items        []ContractorInvoiceLineItem // Individual items
}

// GenerateContractorInvoice generates contractor invoice data from Notion
func (c *controller) GenerateContractorInvoice(ctx context.Context, discord, month string) (*ContractorInvoiceData, error) {
	// Default to current month if not provided
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	l := c.logger.Fields(logger.Fields{
		"discord": discord,
		"month":   month,
	})

	l.Debug("starting contractor invoice generation")

	// 1. Query Contractor Rates
	l.Debug("querying contractor rates from Notion")
	ratesService := notion.NewContractorRatesService(c.config, c.logger)
	if ratesService == nil {
		l.Error(nil, "failed to create contractor rates service")
		return nil, fmt.Errorf("failed to create contractor rates service")
	}

	rateData, err := ratesService.QueryRatesByDiscordAndMonth(ctx, discord, month)
	if err != nil {
		l.Error(err, "failed to query contractor rates")
		return nil, fmt.Errorf("contractor rates not found: %w", err)
	}

	l.Debug(fmt.Sprintf("found contractor rate: contractor=%s billingType=%s currency=%s monthlyFixed=%.2f hourlyRate=%.2f contractorPageID=%s",
		rateData.ContractorName, rateData.BillingType, rateData.Currency, rateData.MonthlyFixed, rateData.HourlyRate, rateData.ContractorPageID))

	// 2. Query Pending Payouts, Refund/Commission before Payday, AND Bank Account in parallel
	l.Debug("[DEBUG] contractor_invoice: starting parallel queries for payouts, refund/commission, and bank account")

	var payouts []notion.PayoutEntry
	var refundCommissionPayouts []notion.PayoutEntry
	var bankAccount *notion.BankAccountData
	var payoutsErr, refundCommissionErr, bankAccountErr error
	var wg sync.WaitGroup

	// Calculate issue date for cutoff date calculation
	payDay := rateData.PayDay
	if payDay <= 0 || payDay > 31 {
		payDay = 1 // Default to 1st if invalid
	}
	monthTime, _ := time.Parse("2006-01", month)
	var issueDate time.Time
	if payDay > 0 && payDay <= 31 {
		// Invoice is issued in the NEXT month on the Payday
		// Example: December 2025 invoice (month=2025-12) issued on January 15th, 2026 (payDay=15)
		nextMonth := monthTime.AddDate(0, 1, 0)
		issueDate = time.Date(nextMonth.Year(), nextMonth.Month(), payDay, 0, 0, 0, 0, time.UTC)
		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: calculated issueDate from Payday: month=%s payDay=%d issueDate=%s", month, payDay, issueDate.Format("2006-01-02")))
	} else {
		// Fallback to current date if Payday is not set
		issueDate = time.Now()
		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: Payday not set, using current date: issueDate=%s", issueDate.Format("2006-01-02")))
	}

	// Build cutoff date for Refund/Commission/Other payouts: use issue date
	cutoffDate := issueDate.Format("2006-01-02")
	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: cutoff date for Refund/Commission/Other payouts: %s (from issueDate)", cutoffDate))

	// Query Payouts (by month)
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.Debug("querying pending payouts from Notion (parallel)")
		payoutsService := notion.NewContractorPayoutsService(c.config, c.logger)
		if payoutsService == nil {
			payoutsErr = fmt.Errorf("failed to create contractor payouts service")
			return
		}
		payouts, payoutsErr = payoutsService.QueryPendingPayoutsByContractor(ctx, rateData.ContractorPageID, month)
		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: payouts query completed, found=%d err=%v", len(payouts), payoutsErr))
	}()

	// Query Refund/Commission/Other payouts before cutoff date
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: querying Refund/Commission/Other payouts before %s (parallel)", cutoffDate))
		payoutsService := notion.NewContractorPayoutsService(c.config, c.logger)
		if payoutsService == nil {
			refundCommissionErr = fmt.Errorf("failed to create contractor payouts service")
			return
		}
		refundCommissionPayouts, refundCommissionErr = payoutsService.QueryPendingRefundCommissionBeforeDate(ctx, rateData.ContractorPageID, cutoffDate)
		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: Refund/Commission/Other query completed, found=%d err=%v", len(refundCommissionPayouts), refundCommissionErr))
	}()

	// Query Bank Account
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.Debug("querying bank account from Notion (parallel)")
		bankAccountService := notion.NewBankAccountService(c.config, c.logger)
		if bankAccountService == nil {
			bankAccountErr = fmt.Errorf("failed to create bank account service")
			return
		}
		bankAccount, bankAccountErr = bankAccountService.QueryBankAccountByDiscord(ctx, discord)
		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: bank account query completed, err=%v", bankAccountErr))
	}()

	wg.Wait()
	l.Debug("[DEBUG] contractor_invoice: parallel queries completed")

	// Check errors
	if payoutsErr != nil {
		l.Error(payoutsErr, "failed to query pending payouts")
		return nil, fmt.Errorf("failed to query pending payouts: %w", payoutsErr)
	}

	// Note: refundCommissionErr is non-fatal - just log and continue
	if refundCommissionErr != nil {
		l.Error(refundCommissionErr, "[DEBUG] contractor_invoice: failed to query Refund/Commission/Other payouts - continuing without them")
	}

	if bankAccountErr != nil {
		l.Error(bankAccountErr, "failed to query bank account")
		return nil, fmt.Errorf("bank account not found: %w", bankAccountErr)
	}

	l.Debug(fmt.Sprintf("found %d pending payouts (month filter), %d Refund/Commission/Other payouts (before payday)", len(payouts), len(refundCommissionPayouts)))

	// Merge Refund/Commission/Other payouts with main payouts (deduplicate by PageID)
	existingPageIDs := make(map[string]bool)
	for _, p := range payouts {
		existingPageIDs[p.PageID] = true
	}
	for _, p := range refundCommissionPayouts {
		if !existingPageIDs[p.PageID] {
			payouts = append(payouts, p)
			existingPageIDs[p.PageID] = true
			l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: merged payout: pageID=%s name=%s type=%s", p.PageID, p.Name, p.SourceType))
		} else {
			l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: skipping duplicate payout: pageID=%s", p.PageID))
		}
	}

	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: total payouts after merge: %d", len(payouts)))

	if len(payouts) == 0 {
		l.Debug("no pending payouts found for contractor")
		return nil, fmt.Errorf("no pending payouts found for contractor=%s", discord)
	}

	// 3. Convert all payout amounts to USD in parallel
	l.Debug("[DEBUG] contractor_invoice: starting parallel currency conversions")
	amountsUSD := make([]float64, len(payouts))
	var convWg sync.WaitGroup
	var convMu sync.Mutex

	for i, payout := range payouts {
		convWg.Add(1)
		go func(idx int, p notion.PayoutEntry) {
			defer convWg.Done()

			amountUSD, _, err := c.service.Wise.Convert(p.Amount, p.Currency, "USD")
			if err != nil {
				convMu.Lock()
				l.Error(err, fmt.Sprintf("failed to convert payout amount to USD: pageID=%s", p.PageID))
				convMu.Unlock()
				amountUSD = p.Amount // Fallback to original amount if conversion fails
			}
			// Round to 2 decimal places
			amountsUSD[idx] = math.Round(amountUSD*100) / 100

			convMu.Lock()
			l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: converted %.2f %s = %.2f USD (idx=%d)", p.Amount, p.Currency, amountsUSD[idx], idx))
			convMu.Unlock()
		}(i, payout)
	}

	convWg.Wait()
	l.Debug("[DEBUG] contractor_invoice: parallel currency conversions completed")

	// Create task order service for hourly rate processing
	taskOrderService := notion.NewTaskOrderLogService(c.config, c.logger)

	// 4. Build line items from payouts
	var lineItems []ContractorInvoiceLineItem
	var total float64

	for i, payout := range payouts {
		amountUSD := amountsUSD[i]

		l.Debug(fmt.Sprintf("processing payout: pageID=%s name=%s sourceType=%s amount=%.2f currency=%s amountUSD=%.2f",
			payout.PageID, payout.Name, payout.SourceType, payout.Amount, payout.Currency, amountUSD))

		// Determine description based on source type
		var description string
		switch payout.SourceType {
		case notion.PayoutSourceTypeServiceFee:
			// For Service Fee: use payout.Description if available, otherwise generate based on positions
			if payout.Description != "" {
				description = payout.Description
				l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: using existing Description for Service Fee: %s", description))
			} else {
				// Generate description based on contractor positions
				// Position contains "design" → "Design Consulting Services Rendered"
				// Position contains "Operation Executive" → "Operational Consulting Services Rendered"
				// Otherwise → "Software Development Services Rendered"
				payoutsService := notion.NewContractorPayoutsService(c.config, c.logger)
				if payoutsService != nil && payout.PersonPageID != "" {
					positions, err := payoutsService.GetContractorPositions(ctx, payout.PersonPageID)
					if err != nil {
						l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: failed to fetch contractor positions: %v", err))
					}
					description = generateServiceFeeDescription(month, positions)
					l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: generated Service Fee description from positions: %s", description))
				} else {
					description = generateServiceFeeDescription(month, nil)
					l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: using default Service Fee description (no positions): %s", description))
				}
			}
			// Final fallback: generate default description from invoice month
			if description == "" {
				description = generateDefaultServiceFeeDescription(month)
				l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: generating default description from month: %s", description))
			}
		case notion.PayoutSourceTypeRefund:
			// For Refund: use Description field, fallback to Name if empty
			if payout.Description != "" {
				description = payout.Description
				l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: using Description for Refund: length=%d", len(description)))
			} else {
				description = payout.Name
				l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: Description empty, using Name for Refund: %s", payout.Name))
			}
		default:
			// Commission and Other: use Description field
			description = payout.Description
			l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: using Description field: length=%d", len(description)))
		}

		// New: Attempt hourly rate display for Service Fees
		var lineItem ContractorInvoiceLineItem
		isHourlyProcessed := false

		if payout.SourceType == notion.PayoutSourceTypeServiceFee && ratesService != nil && taskOrderService != nil {
			hourlyData := fetchHourlyRateData(ctx, payout, ratesService, taskOrderService, l)
			if hourlyData != nil {
				// SUCCESS: Use hourly rate display
				l.Debug(fmt.Sprintf("[SUCCESS] payout %s: applying hourly rate display (hours=%.2f rate=%.2f)",
					payout.PageID, hourlyData.Hours, hourlyData.HourlyRate))

				lineItem = ContractorInvoiceLineItem{
					Title:             payout.Description, // Use payout Description as section title
					Description:       description,
					Hours:             hourlyData.Hours,
					Rate:              hourlyData.HourlyRate,
					Amount:            amountUSD, // Use USD-converted amount
					AmountUSD:         amountUSD,
					Type:              string(payout.SourceType),
					CommissionRole:    payout.CommissionRole,
					CommissionProject: payout.CommissionProject,
					OriginalAmount:    payout.Amount,
					OriginalCurrency:  payout.Currency,
					IsHourlyRate:      true,
					ServiceRateID:     payout.ServiceRateID,
					TaskOrderID:       payout.TaskOrderID,
				}
				isHourlyProcessed = true
			}
		}

		if !isHourlyProcessed {
			// FALLBACK / STANDARD: Use default display
			if payout.SourceType == notion.PayoutSourceTypeServiceFee {
				l.Debug(fmt.Sprintf("[FALLBACK] payout %s: using default display (Qty=1, Unit Cost=%.2f)",
					payout.PageID, amountUSD))
			} else {
				l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: %s - no title, only description", payout.SourceType))
			}

			// Set Title from payout.Description for Service Fee items (used as section name)
			title := ""
			if payout.SourceType == notion.PayoutSourceTypeServiceFee {
				title = payout.Description
			}

			lineItem = ContractorInvoiceLineItem{
				Title:             title,
				Description:       description,
				Hours:             1,             // Default quantity
				Rate:              amountUSD, // Unit cost in USD
				Amount:            amountUSD, // Amount in USD
				AmountUSD:         amountUSD,
				Type:              string(payout.SourceType),
				CommissionRole:    payout.CommissionRole,
				CommissionProject: payout.CommissionProject,
				// Preserve original currency for multi-currency support
				OriginalAmount:   payout.Amount,
				OriginalCurrency: payout.Currency,
				IsHourlyRate:     false,
				ServiceRateID:    payout.ServiceRateID,
				TaskOrderID:      payout.TaskOrderID,
			}
		}

		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: preserved original currency - %.2f %s (converted to %.2f USD)",
			payout.Amount, payout.Currency, amountUSD))

		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: line item - Type=%s Role=%s Project=%s",
			lineItem.Type, lineItem.CommissionRole, lineItem.CommissionProject))

		lineItems = append(lineItems, lineItem)
		total += amountUSD
	}

	l.Debug(fmt.Sprintf("built %d line items with total=%.2f USD", len(lineItems), total))

	// 4.5 Aggregate hourly-rate Service Fees
	l.Debug("aggregating hourly-rate Service Fee items")
	lineItems = aggregateHourlyServiceFees(lineItems, month, l)
	l.Debug(fmt.Sprintf("after aggregation: %d line items", len(lineItems)))

	// 4.6 Group Commission items by Project (all commissions for same project are summed)
	l.Debug("grouping commission items by project")
	var nonCommissionItems []ContractorInvoiceLineItem
	commissionGroups := make(map[string]float64) // key = project name

	for _, item := range lineItems {
		if item.Type != string(notion.PayoutSourceTypeCommission) {
			nonCommissionItems = append(nonCommissionItems, item)
			continue
		}

		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: commission - project=%s amount=%.2f",
			item.CommissionProject, item.AmountUSD))

		// Group by project name (empty string if no project)
		commissionGroups[item.CommissionProject] += item.AmountUSD
	}

	// Convert grouped commissions to line items
	var groupedCommissionItems []ContractorInvoiceLineItem
	for project, totalAmount := range commissionGroups {
		// Round total to 2 decimal places
		groupTotal := math.Round(totalAmount*100) / 100

		// Build description: "Fee for Renaiss" or "Fee"
		description := "Fee"
		if project != "" {
			description = fmt.Sprintf("Fee for %s", project)
		}

		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: grouped commission - description=%s total=%.2f", description, groupTotal))

		groupedCommissionItems = append(groupedCommissionItems, ContractorInvoiceLineItem{
			Title:       "",
			Description: description,
			Hours:       1,
			Rate:        groupTotal,
			Amount:      groupTotal,
			AmountUSD:   groupTotal,
			Type:        string(notion.PayoutSourceTypeCommission),
			// Grouped commissions are in USD (sum of converted amounts)
			OriginalAmount:   groupTotal,
			OriginalCurrency: "USD",
		})

		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: set grouped commission currency - %.2f USD", groupTotal))
	}

	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: grouped %d commission items into %d groups",
		len(lineItems)-len(nonCommissionItems), len(groupedCommissionItems)))

	// Combine non-commission items with grouped commission items
	lineItems = append(nonCommissionItems, groupedCommissionItems...)

	// 5. Sort line items by Type (Service Fee last) then by Amount ASC
	l.Debug("sorting line items by Type (Service Fee last) and Amount ASC")
	sort.Slice(lineItems, func(i, j int) bool {
		// Service Fee should always be last
		iIsServiceFee := lineItems[i].Type == string(notion.PayoutSourceTypeServiceFee)
		jIsServiceFee := lineItems[j].Type == string(notion.PayoutSourceTypeServiceFee)
		if iIsServiceFee != jIsServiceFee {
			return !iIsServiceFee // Service Fee goes to the end
		}
		// Then sort by Type (group by type)
		if lineItems[i].Type != lineItems[j].Type {
			return lineItems[i].Type < lineItems[j].Type
		}
		// Then sort by Amount ASC within each type
		return lineItems[i].Amount < lineItems[j].Amount
	})
	l.Debug(fmt.Sprintf("sorted %d line items by Type (Service Fee last) and Amount ASC", len(lineItems)))

	// 5.5 Validate currencies before calculations
	l.Debug("[DEBUG] contractor_invoice: validating line item currencies")
	if err := validateLineItemCurrencies(lineItems, l); err != nil {
		l.Error(err, "contractor_invoice: currency validation failed")
		return nil, err
	}

	// 5.55 Fetch exchange rate for display (USD → VND)
	var exchangeRate float64
	hasVNDItems := false
	for _, item := range lineItems {
		if item.OriginalCurrency == "VND" {
			hasVNDItems = true
			break
		}
	}
	if hasVNDItems {
		l.Debug("[DEBUG] contractor_invoice: fetching exchange rate for display (USD → VND)")
		quote, err := c.service.Wise.GetPayrollQuotes("USD", "VND", 1) // Get rate for $1
		if err != nil {
			l.Warn(fmt.Sprintf("contractor_invoice: failed to get exchange rate: %v, will not show FX rate in invoice", err))
		} else {
			exchangeRate = quote.Rate
			l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: exchange rate for display: %.4f", exchangeRate))
		}
	}

	// 5.6 Calculate subtotal by summing AmountUSD from all line items
	var subtotalUSD float64
	for _, item := range lineItems {
		subtotalUSD += item.AmountUSD
		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: item AmountUSD: %.2f (running total: %.2f)", item.AmountUSD, subtotalUSD))
	}

	// Round to 2 decimal places
	subtotalUSD = math.Round(subtotalUSD*100) / 100
	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: subtotal USD: %.2f", subtotalUSD))

	// 5.7 Calculate final total (same as subtotal since FX support removed)
	totalUSD := subtotalUSD

	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: final total USD: %.2f", totalUSD))

	l.Debug(fmt.Sprintf("found bank account: accountHolder=%s bank=%s accountNumber=%s",
		bankAccount.AccountHolderName, bankAccount.BankName, bankAccount.AccountNumber))

	// 7. Generate invoice number
	invoiceNumber := c.generateContractorInvoiceNumber(month)
	l.Debug(fmt.Sprintf("generated invoice number: %s", invoiceNumber))

	// 8. Calculate due date (issueDate already calculated earlier for cutoff date)
	dueDate := issueDate.AddDate(0, 0, 9) // Due in 9 days from issue date (Payday + 9)

	// 9. Generate description
	description := fmt.Sprintf("Professional Services for %s", timeutil.FormatMonthYear(month))
	l.Debug(fmt.Sprintf("generated description: %s", description))

	invoiceData := &ContractorInvoiceData{
		InvoiceNumber:     invoiceNumber,
		ContractorName:    rateData.ContractorName,
		Month:             month,
		Date:              issueDate,
		DueDate:           dueDate,
		Description:       description,
		BillingType:       rateData.BillingType,
		Currency:          "USD", // Invoice currency for payment is always USD
		LineItems:         lineItems,
		MonthlyFixed:      0,
		MonthlyFixedUSD:   0,
		Total:             totalUSD,     // Use calculated total
		TotalUSD:          totalUSD,     // Use calculated total
		ExchangeRate:      exchangeRate, // For FX rate footnote display
		BankAccountHolder: bankAccount.AccountHolderName,
		BankName:          bankAccount.BankName,
		BankAccountNumber: bankAccount.AccountNumber,
		BankSwiftBIC:      bankAccount.SwiftBIC,
		BankBranch:        bankAccount.BranchAddress,

		// Populate subtotal fields
		SubtotalVND:        0,           // No longer used (simplified calculation)
		SubtotalUSDFromVND: 0,           // Deprecated - no longer used
		SubtotalUSDItems:   0,           // Deprecated - no longer used
		SubtotalUSD:        subtotalUSD, // Direct sum of AmountUSD from line items

		// Notion relation IDs (for creating Contractor Payables record)
		ContractorPageID: rateData.ContractorPageID,

		// PayDay for Period calculation
		PayDay: payDay,
	}

	// Collect payout page IDs from processed payouts
	payoutPageIDs := make([]string, len(payouts))
	for i, payout := range payouts {
		payoutPageIDs[i] = payout.PageID
	}
	invoiceData.PayoutPageIDs = payoutPageIDs
	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: collected %d payout page IDs for Contractor Payables record", len(payoutPageIDs)))

	l.Debug("[DEBUG] contractor_invoice: invoice data populated with calculated totals")
	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: SubtotalUSD=%.2f TotalUSD=%.2f ExchangeRate=%.4f",
		subtotalUSD, totalUSD, exchangeRate))
	return invoiceData, nil
}

// generateContractorInvoiceNumber generates invoice number in format CONTR-{YYYYMM}-{random-4-chars}
func (c *controller) generateContractorInvoiceNumber(month string) string {
	// Remove hyphen from month (2025-12 -> 202512)
	monthPart := strings.ReplaceAll(month, "-", "")
	randomChars := generateRandomAlphanumeric(4)
	return fmt.Sprintf("CONTR-%s-%s", monthPart, randomChars)
}

// generateRandomAlphanumeric generates a random alphanumeric string of given length
func generateRandomAlphanumeric(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to fixed value if random fails
		return "A1B2"
	}
	for i := range result {
		result[i] = charset[int(randomBytes[i])%len(charset)]
	}
	return string(result)
}

// GenerateContractorInvoicePDF generates PDF from contractor invoice data
func (c *controller) GenerateContractorInvoicePDF(l logger.Logger, data *ContractorInvoiceData) ([]byte, error) {
	l.Debug(fmt.Sprintf("generating contractor invoice PDF: invoiceNumber=%s billingType=%s total=%.2f",
		data.InvoiceNumber, data.BillingType, data.Total))

	// Setup currency formatter
	currencyCode := data.Currency
	if currencyCode == "" {
		currencyCode = "VND"
	}
	pound := money.New(1, currencyCode)

	l.Debug(fmt.Sprintf("using currency: %s", currencyCode))

	// Create template FuncMap
	funcMap := template.FuncMap{
		"formatMoney": func(amount float64) string {
			tmpValue := amount * math.Pow(10, float64(pound.Currency().Fraction))
			// Use math.Round to ensure consistent rounding (not truncation)
			return pound.Multiply(int64(math.Round(tmpValue))).Display()
		},
		"formatDate": func(t time.Time) string {
			return timeutil.FormatDatetime(t)
		},
		"isMonthlyFixed": func() bool {
			return data.BillingType == "Monthly Fixed"
		},
		"isHourlyRate": func() bool {
			return data.BillingType == "Hourly Rate"
		},
		"add": func(a, b int) int {
			return a + b
		},
		"float": func(n float64) string {
			return fmt.Sprintf("%.2f", n)
		},
		"formatProofOfWork": func(text string) template.HTML {
			l.Debug(fmt.Sprintf("[DEBUG] formatProofOfWork INPUT (len=%d): %q", len(text), text))

			// Replace bullet points with newlines for better formatting
			formatted := strings.ReplaceAll(text, " • ", "\n• ")
			formatted = strings.ReplaceAll(formatted, " •", "\n•")

			// Trim whitespace first
			formatted = strings.TrimSpace(formatted)

			// Replace double newlines with a spacer div for controlled spacing between projects
			formatted = strings.ReplaceAll(formatted, "\n\n", "\n<div class=\"project-spacer\"></div>")

			// Replace newlines with <br> for HTML rendering
			formatted = strings.ReplaceAll(formatted, "\n", "<br>")

			// Remove ALL trailing <br> tags
			for strings.HasSuffix(formatted, "<br>") {
				formatted = strings.TrimSuffix(formatted, "<br>")
			}
			// Remove ALL leading <br> tags
			for strings.HasPrefix(formatted, "<br>") {
				formatted = strings.TrimPrefix(formatted, "<br>")
			}
			l.Debug(fmt.Sprintf("[DEBUG] formatProofOfWork OUTPUT (len=%d): %q", len(formatted), formatted))

			return template.HTML(formatted)
		},
		// Multi-currency formatting functions (added for multi-currency support)
		"formatCurrency":     formatCurrency,
		"formatVND":          formatVND,
		"formatUSD":          formatUSD,
		"formatExchangeRate": formatExchangeRate,
	}

	// Determine template path
	templatePath := c.config.Invoice.TemplatePath
	if c.config.Env == "local" {
		cwd, err := os.Getwd()
		if err == nil {
			templatePath = filepath.Join(cwd, "pkg/templates")
			l.Debug(fmt.Sprintf("[Local Env] Using template path from cwd: %s", templatePath))
		} else {
			l.Debug(fmt.Sprintf("[Local Env] Failed to get cwd, using config path: %s", templatePath))
		}
	}

	templateFile := filepath.Join(templatePath, "contractor-invoice-template.html")
	l.Debug(fmt.Sprintf("parsing template from: %s", templateFile))

	// Parse template
	tmpl, err := template.New("contractorInvoicePDF").Funcs(funcMap).ParseFiles(templateFile)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to parse template: %s", templateFile))
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// All line items are shown as regular items (no merging for payout-based invoices)
	regularItems := data.LineItems
	var mergedItems []ContractorInvoiceLineItem
	mergedTotal := 0.0

	l.Debug(fmt.Sprintf("line items: %d regular, %d merged", len(regularItems), len(mergedItems)))

	// Group line items into sections (Development Work, Refund, Bonus)
	sections := groupLineItemsIntoSections(regularItems, l)

	// Prepare template data
	templateData := struct {
		Invoice     *ContractorInvoiceData
		LineItems   []ContractorInvoiceLineItem
		MergedItems []ContractorInvoiceLineItem
		MergedTotal float64
		Sections    []ContractorInvoiceSection
	}{
		Invoice:     data,
		LineItems:   regularItems,
		MergedItems: mergedItems,
		MergedTotal: mergedTotal,
		Sections:    sections,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "contractor-invoice-template.html", templateData); err != nil {
		l.Error(err, "failed to execute template")
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	l.Debug("template executed successfully")

	// Convert HTML to PDF using wkhtmltopdf
	pdfg, err := toPdf.NewPDFGenerator()
	if err != nil {
		l.Error(err, "failed to create pdf generator")
		return nil, fmt.Errorf("failed to create pdf generator: %w", err)
	}

	page := toPdf.NewPageReader(&buf)
	page.Zoom.Set(1.0)
	page.EnableLocalFileAccess.Set(true)
	pdfg.AddPage(page)
	pdfg.Dpi.Set(600)
	pdfg.PageSize.Set("A4")

	if err := pdfg.Create(); err != nil {
		l.Error(err, "failed to create PDF")
		return nil, fmt.Errorf("failed to create PDF: %w", err)
	}

	pdfBytes := pdfg.Buffer().Bytes()
	l.Debug(fmt.Sprintf("PDF generated successfully: size=%d bytes", len(pdfBytes)))

	return pdfBytes, nil
}

// validateLineItemCurrencies validates that all line items have valid currencies and amounts
func validateLineItemCurrencies(lineItems []ContractorInvoiceLineItem, l logger.Logger) error {
	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: validating currencies for %d line items", len(lineItems)))

	for i, item := range lineItems {
		// Validate currency code
		if item.OriginalCurrency != "VND" && item.OriginalCurrency != "USD" {
			l.Error(nil, fmt.Sprintf("contractor_invoice: invalid currency at item %d: %s", i, item.OriginalCurrency))
			return fmt.Errorf("invalid currency for line item %d: %s (must be VND or USD)", i, item.OriginalCurrency)
		}

		// Validate amount is non-negative
		if item.OriginalAmount < 0 {
			l.Error(nil, fmt.Sprintf("contractor_invoice: negative amount at item %d: %.2f %s", i, item.OriginalAmount, item.OriginalCurrency))
			return fmt.Errorf("negative amount for line item %d: %.2f %s", i, item.OriginalAmount, item.OriginalCurrency)
		}

		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: validated item %d: %.2f %s", i, item.OriginalAmount, item.OriginalCurrency))
	}

	l.Debug("[DEBUG] contractor_invoice: all line item currencies validated successfully")
	return nil
}

// formatVND formats a VND amount with Vietnamese conventions
// Example: 45000000 -> "45.000.000 ₫"
func formatVND(amount float64) string {
	// Round to 0 decimal places (VND has no minor units)
	rounded := math.Round(amount)

	// Convert to string without decimals
	str := fmt.Sprintf("%.0f", rounded)

	// Handle negative numbers
	negative := false
	if str[0] == '-' {
		negative = true
		str = str[1:]
	}

	// Add period separators for thousands
	var result []rune
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, char)
	}

	// Add currency symbol after amount
	formatted := string(result) + " ₫"
	if negative {
		formatted = "-" + formatted
	}

	return formatted
}

// formatUSD formats a USD amount with US conventions
// Example: 1234.56 -> "$1,234.56"
func formatUSD(amount float64) string {
	// Round to 2 decimal places
	rounded := math.Round(amount*100) / 100

	// Handle negative numbers
	negative := false
	if rounded < 0 {
		negative = true
		rounded = -rounded
	}

	// Split into integer and decimal parts
	intPart := int64(rounded)
	decPart := int((rounded - float64(intPart)) * 100)

	// Format integer part with comma separators
	intStr := fmt.Sprintf("%d", intPart)
	var result []rune
	for i, char := range intStr {
		if i > 0 && (len(intStr)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, char)
	}

	// Add $ symbol before amount and decimal part
	formatted := fmt.Sprintf("$%s.%02d", string(result), decPart)
	if negative {
		formatted = "-" + formatted
	}

	return formatted
}

// formatCurrency formats an amount according to its currency code
func formatCurrency(amount float64, currency string) string {
	switch strings.ToUpper(currency) {
	case "VND":
		return formatVND(amount)
	case "USD":
		return formatUSD(amount)
	default:
		// Fallback to USD formatting for unknown currencies
		return formatUSD(amount)
	}
}

// formatExchangeRate formats an exchange rate for display
// Example: 26269.123 -> "1 USD = 26,269 VND"
func formatExchangeRate(rate float64) string {
	// Round to nearest whole number for VND
	rounded := math.Round(rate)

	// Format with comma separators
	intStr := fmt.Sprintf("%.0f", rounded)
	var result []rune
	for i, char := range intStr {
		if i > 0 && (len(intStr)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, char)
	}

	return fmt.Sprintf("1 USD = %s VND", string(result))
}

// groupLineItemsIntoSections groups line items into sections for display
// Sections: Development Work (Service Fee) - aggregated with total, Refund - individual items, Bonus (Commission) - individual items
func groupLineItemsIntoSections(items []ContractorInvoiceLineItem, l logger.Logger) []ContractorInvoiceSection {
	var sections []ContractorInvoiceSection

	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: grouping %d line items into sections", len(items)))

	// Group Service Fee items (Development Work) - aggregated display
	var serviceFeeItems []ContractorInvoiceLineItem
	for _, item := range items {
		// Service Fee items are contractor payroll/work-related payouts
		if item.Type == string(notion.PayoutSourceTypeServiceFee) {
			serviceFeeItems = append(serviceFeeItems, item)
		}
	}

	if len(serviceFeeItems) > 0 {
		// Calculate total in USD
		totalUSD := 0.0
		for _, item := range serviceFeeItems {
			totalUSD += item.AmountUSD
		}

		// Round total to 2 decimal places (USD)
		totalUSD = math.Round(totalUSD*100) / 100

		// Use Title from first service fee item as section name (from payout.Description)
		sectionName := serviceFeeItems[0].Title

		sections = append(sections, ContractorInvoiceSection{
			Name:         sectionName,
			IsAggregated: true,
			Total:        totalUSD,
			Currency:     "USD",
			Items:        serviceFeeItems,
		})

		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: created Development Work section with %d items, total: %.2f USD, name: %s",
			len(serviceFeeItems), totalUSD, sectionName))
	}

	// Group Refund items - individual display
	var refundItems []ContractorInvoiceLineItem
	for _, item := range items {
		if item.Type == string(notion.PayoutSourceTypeRefund) {
			refundItems = append(refundItems, item)
		}
	}

	if len(refundItems) > 0 {
		sections = append(sections, ContractorInvoiceSection{
			Name:         "Refund",
			IsAggregated: false,
			Items:        refundItems,
		})

		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: created Refund section with %d items", len(refundItems)))
	}

	// Group Bonus (Commission) items - individual display
	var feeItems []ContractorInvoiceLineItem
	for i, item := range items {
		items[i].Description = strings.Replace(items[i].Description, "Bonus", "Fee", -1)
		if item.Type == string(notion.PayoutSourceTypeCommission) {
			feeItems = append(feeItems, item)
		}
	}

	if len(feeItems) > 0 {
		sections = append(sections, ContractorInvoiceSection{
			Name:         "Fee",
			IsAggregated: false,
			Items:        feeItems,
		})

		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: created Fee section with %d items", len(feeItems)))
	}

	// Group Extra Payment items - individual display
	var extraPaymentItems []ContractorInvoiceLineItem
	for _, item := range items {
		if item.Type == string(notion.PayoutSourceTypeExtraPayment) {
			extraPaymentItems = append(extraPaymentItems, item)
		}
	}

	if len(extraPaymentItems) > 0 {
		sections = append(sections, ContractorInvoiceSection{
			Name:         "Extra Payment",
			IsAggregated: false,
			Items:        extraPaymentItems,
		})

		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: created Extra Payment section with %d items", len(extraPaymentItems)))
	}

	l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: grouped into %d sections", len(sections)))

	return sections
}

// hourlyRateData holds fetched data for hourly rate Service Fee display.
type hourlyRateData struct {
	HourlyRate    float64
	Hours         float64
	Currency      string
	BillingType   string
	ServiceRateID string
	TaskOrderID   string
}

// hourlyRateAggregation holds aggregated data for multiple hourly Service Fees.
type hourlyRateAggregation struct {
	TotalHours     float64
	HourlyRate     float64
	TotalAmount    float64
	TotalAmountUSD float64
	Currency       string
	Descriptions   []string
	TaskOrderIDs   []string
}

// IContractorRatesService defines the interface for contractor rates operations needed by the invoice controller
type IContractorRatesService interface {
	FetchContractorRateByPageID(ctx context.Context, pageID string) (*notion.ContractorRateData, error)
}

// ITaskOrderLogService defines the interface for task order log operations needed by the invoice controller
type ITaskOrderLogService interface {
	FetchTaskOrderHoursByPageID(ctx context.Context, pageID string) (float64, error)
}

// fetchHourlyRateData fetches and validates hourly rate data for a Service Fee payout.
func fetchHourlyRateData(
	ctx context.Context,
	payout notion.PayoutEntry,
	ratesService IContractorRatesService,
	taskOrderService ITaskOrderLogService,
	l logger.Logger,
) *hourlyRateData {
	// STEP 1: Check ServiceRateID present
	if payout.ServiceRateID == "" {
		l.Debug(fmt.Sprintf("[FALLBACK] payout %s: no ServiceRateID", payout.PageID))
		return nil
	}

	// STEP 2: Fetch Contractor Rate
	l.Debug(fmt.Sprintf("[HOURLY_RATE] fetching contractor rate: serviceRateID=%s", payout.ServiceRateID))
	rateData, err := ratesService.FetchContractorRateByPageID(ctx, payout.ServiceRateID)
	if err != nil {
		l.Error(err, fmt.Sprintf("[FALLBACK] payout %s: failed to fetch rate", payout.PageID))
		return nil
	}

	l.Debug(fmt.Sprintf("[HOURLY_RATE] fetched rate: billingType=%s hourlyRate=%.2f currency=%s",
		rateData.BillingType, rateData.HourlyRate, rateData.Currency))

	// STEP 3: Validate BillingType
	if rateData.BillingType != "Hourly Rate" {
		l.Debug(fmt.Sprintf("[INFO] payout %s: billingType=%s (not hourly)", payout.PageID, rateData.BillingType))
		return nil
	}

	// STEP 4: Fetch Task Order hours (graceful degradation)
	var hours float64
	if payout.TaskOrderID != "" {
		l.Debug(fmt.Sprintf("[HOURLY_RATE] fetching hours: taskOrderID=%s", payout.TaskOrderID))
		hours, err = taskOrderService.FetchTaskOrderHoursByPageID(ctx, payout.TaskOrderID)
		if err != nil {
			l.Error(err, fmt.Sprintf("[FALLBACK] payout %s: failed to fetch hours, using 0", payout.PageID))
			hours = 0
		} else {
			l.Debug(fmt.Sprintf("[HOURLY_RATE] fetched hours: %.2f", hours))
		}
	} else {
		l.Debug(fmt.Sprintf("[FALLBACK] payout %s: no TaskOrderID, using 0 hours", payout.PageID))
		hours = 0
	}

	// STEP 5: Create hourlyRateData
	return &hourlyRateData{
		HourlyRate:    rateData.HourlyRate,
		Hours:         hours,
		Currency:      rateData.Currency,
		BillingType:   rateData.BillingType,
		ServiceRateID: payout.ServiceRateID,
		TaskOrderID:   payout.TaskOrderID,
	}
}

// aggregateHourlyServiceFees consolidates all hourly-rate Service Fee items into a single line item.
func aggregateHourlyServiceFees(
	lineItems []ContractorInvoiceLineItem,
	month string,
	l logger.Logger,
) []ContractorInvoiceLineItem {
	// STEP 1: Partition line items
	var hourlyItems []ContractorInvoiceLineItem
	var otherItems []ContractorInvoiceLineItem

	for _, item := range lineItems {
		if item.IsHourlyRate {
			hourlyItems = append(hourlyItems, item)
		} else {
			otherItems = append(otherItems, item)
		}
	}

	l.Debug(fmt.Sprintf("[AGGREGATE] found %d hourly-rate Service Fee items to aggregate", len(hourlyItems)))

	if len(hourlyItems) == 0 {
		l.Debug("[AGGREGATE] no hourly items, returning unchanged")
		return lineItems
	}

	// STEP 2: Aggregate hourly items
	agg := &hourlyRateAggregation{}

	for _, item := range hourlyItems {
		// Sum hours and amounts
		agg.TotalHours += item.Hours
		agg.TotalAmount += item.Amount
		agg.TotalAmountUSD += item.AmountUSD

		// Use first item's rate and currency
		if agg.HourlyRate == 0 {
			agg.HourlyRate = item.Rate
			agg.Currency = item.OriginalCurrency
		} else {
			// Validate consistency (log warnings)
			if item.Rate != agg.HourlyRate {
				l.Warn(fmt.Sprintf("[WARN] multiple hourly rates found: %.2f vs %.2f, using first",
					agg.HourlyRate, item.Rate))
			}
			if item.OriginalCurrency != agg.Currency {
				l.Warn(fmt.Sprintf("[WARN] multiple currencies found: %s vs %s, using first",
					agg.Currency, item.OriginalCurrency))
			}
		}

		// Collect descriptions
		if strings.TrimSpace(item.Description) != "" {
			agg.Descriptions = append(agg.Descriptions, strings.TrimSpace(item.Description))
		}

		// Collect task order IDs (for logging)
		if item.TaskOrderID != "" {
			agg.TaskOrderIDs = append(agg.TaskOrderIDs, item.TaskOrderID)
		}
	}

	l.Debug(fmt.Sprintf("[AGGREGATE] totalHours=%.2f rate=%.2f totalAmount=%.2f totalAmountUSD=%.2f currency=%s",
		agg.TotalHours, agg.HourlyRate, agg.TotalAmount, agg.TotalAmountUSD, agg.Currency))

	// STEP 3: Generate title
	title := generateServiceFeeTitle(month)

	// STEP 4: Concatenate descriptions
	description := concatenateDescriptions(agg.Descriptions)

	// STEP 5: Create aggregated line item
	aggregatedItem := ContractorInvoiceLineItem{
		Title:            title,
		Description:      description,
		Hours:            agg.TotalHours,
		Rate:             agg.HourlyRate,
		Amount:           agg.TotalAmount,
		AmountUSD:        agg.TotalAmountUSD,
		Type:             string(notion.PayoutSourceTypeServiceFee), // Use ServiceFee type
		OriginalAmount:   agg.TotalAmount,
		OriginalCurrency: agg.Currency,
		IsHourlyRate:     false, // Already aggregated
		ServiceRateID:    "",
		TaskOrderID:      "",
	}

	l.Debug(fmt.Sprintf("[AGGREGATE] created aggregated item with title: %s", title))
	l.Debug(fmt.Sprintf("[AGGREGATE] aggregated %d items from task orders: %v",
		len(hourlyItems), agg.TaskOrderIDs))

	return append(otherItems, aggregatedItem)
}

// generateServiceFeeTitle generates title with invoice month date range.
func generateServiceFeeTitle(month string) string {
	// STEP 1: Parse month
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return "Service Fee" // Fallback
	}

	// STEP 2: Calculate date range
	startDate := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	// STEP 3: Format title
	return fmt.Sprintf("Service Fee (Development work from %s to %s)",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))
}

// generateServiceFeeDescription generates service fee description based on contractor positions.
// Priority: design > operation executive > default (software development)
// - Position contains "design" → "Design Consulting Services Rendered (Month Day-Day, Year)"
// - Position contains "Operation Executive" → "Operational Consulting Services Rendered (Month Day-Day, Year)"
// - Otherwise → "Software Development Services Rendered (Month Day-Day, Year)"
func generateServiceFeeDescription(month string, positions []string) string {
	// STEP 1: Parse month and calculate date range
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return "Software Development Services Rendered"
	}

	startDate := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	dateRange := fmt.Sprintf("(%s %d-%d, %d)",
		startDate.Format("January"), startDate.Day(), endDate.Day(), startDate.Year())

	// STEP 2: Check positions with priority (case-insensitive)
	// Priority: design > operation executive > default
	hasDesign := false
	hasOperationExecutive := false

	for _, pos := range positions {
		posLower := strings.ToLower(pos)
		if strings.Contains(posLower, "design") {
			hasDesign = true
		}
		if strings.Contains(posLower, "operation executive") {
			hasOperationExecutive = true
		}
	}

	// Return based on priority
	if hasDesign {
		return "Design Consulting Services Rendered " + dateRange
	}
	if hasOperationExecutive {
		return "Operational Consulting Services Rendered " + dateRange
	}

	// STEP 3: Default to software development
	return "Software Development Services Rendered " + dateRange
}

// concatenateDescriptions joins descriptions with double line breaks, filtering empty strings.
func concatenateDescriptions(descriptions []string) string {
	// STEP 1: Filter empty strings
	var filtered []string
	for _, desc := range descriptions {
		trimmed := strings.TrimSpace(desc)
		if trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}

	// STEP 2: Join with double line breaks
	return strings.Join(filtered, "\n\n")
}

// generateDefaultServiceFeeDescription creates a default description based on invoice month.
// Input format: "2006-01" (e.g., "2025-01")
// Output format: "Professional work from Jan 1 to Jan 31"
func generateDefaultServiceFeeDescription(month string) string {
	if month == "" {
		return ""
	}

	t, err := time.Parse("2006-01", month)
	if err != nil {
		return ""
	}

	// Get first and last day of the month
	startDate := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1) // Last day of month

	return fmt.Sprintf("Professional work from %s %d to %s %d",
		startDate.Format("Jan"), startDate.Day(),
		endDate.Format("Jan"), endDate.Day())
}
