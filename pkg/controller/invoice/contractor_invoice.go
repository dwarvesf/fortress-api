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

	// 2. Query Pending Payouts AND Bank Account in parallel
	l.Debug("[DEBUG] contractor_invoice: starting parallel queries for payouts and bank account")

	var payouts []notion.PayoutEntry
	var bankAccount *notion.BankAccountData
	var payoutsErr, bankAccountErr error
	var wg sync.WaitGroup

	// Query Payouts
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.Debug("querying pending payouts from Notion (parallel)")
		payoutsService := notion.NewContractorPayoutsService(c.config, c.logger)
		if payoutsService == nil {
			payoutsErr = fmt.Errorf("failed to create contractor payouts service")
			return
		}
		payouts, payoutsErr = payoutsService.QueryPendingPayoutsByContractor(ctx, rateData.ContractorPageID)
		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: payouts query completed, found=%d err=%v", len(payouts), payoutsErr))
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

	if len(payouts) == 0 {
		l.Debug("no pending payouts found for contractor")
		return nil, fmt.Errorf("no pending payouts found for contractor=%s", discord)
	}

	if bankAccountErr != nil {
		l.Error(bankAccountErr, "failed to query bank account")
		return nil, fmt.Errorf("bank account not found: %w", bankAccountErr)
	}

	l.Debug(fmt.Sprintf("found %d pending payouts", len(payouts)))

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

	// 4. Build line items from payouts
	var lineItems []ContractorInvoiceLineItem
	var total float64

	for i, payout := range payouts {
		amountUSD := amountsUSD[i]

		l.Debug(fmt.Sprintf("processing payout: pageID=%s name=%s sourceType=%s amount=%.2f currency=%s amountUSD=%.2f",
			payout.PageID, payout.Name, payout.SourceType, payout.Amount, payout.Currency, amountUSD))

		// Determine description based on source type:
		// - Service Fee: use WorkDetails from Notion formula
		// - Refund: use Description field as title (displayed in Description column)
		// - Commission: use Description field (will be parsed and grouped later)
		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: payout fields - Name=%s Description=%s WorkDetails=%s",
			payout.Name, payout.Description, payout.WorkDetails))

		var description string
		switch payout.SourceType {
		case notion.PayoutSourceTypeServiceFee:
			// For Service Fee: fetch proof of works from Task Order Log subitems
			if payout.TaskOrderID != "" {
				l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: fetching proof of works from Task Order Log: taskOrderID=%s", payout.TaskOrderID))
				formattedPOW, err := c.service.Notion.TaskOrderLog.FormatProofOfWorksByProject(ctx, []string{payout.TaskOrderID})
				if err != nil {
					l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: failed to fetch proof of works from Task Order Log: %v", err))
				} else if formattedPOW != "" {
					description = formattedPOW
					l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: using formatted proof of works from Task Order Log: length=%d", len(description)))
				}
			}
			// Fallback to WorkDetails if Task Order lookup failed or returned empty
			if description == "" && payout.WorkDetails != "" {
				description = payout.WorkDetails
				l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: using WorkDetails for Service Fee: length=%d", len(description)))
			} else if description == "" {
				description = payout.Description
				l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: WorkDetails empty, using Description for Service Fee: length=%d", len(description)))
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

		// Initialize line item with default values for display
		// All items: Qty=1, Unit Cost=AmountUSD, Total=AmountUSD
		// Title: All types use empty title, only description is shown
		title := ""
		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: %s - no title, only description", payout.SourceType))

		lineItem := ContractorInvoiceLineItem{
			Title:             title,
			Description:       description,
			Hours:             1,         // Default quantity
			Rate:              amountUSD, // Unit cost = converted amount
			Amount:            amountUSD, // Amount
			AmountUSD:         amountUSD,
			Type:              string(payout.SourceType),
			CommissionRole:    payout.CommissionRole,
			CommissionProject: payout.CommissionProject,
		}

		l.Debug(fmt.Sprintf("[DEBUG] contractor_invoice: line item - Type=%s Role=%s Project=%s",
			lineItem.Type, lineItem.CommissionRole, lineItem.CommissionProject))

		lineItems = append(lineItems, lineItem)
		total += amountUSD
	}

	l.Debug(fmt.Sprintf("built %d line items with total=%.2f USD", len(lineItems), total))

	// 4.5 Group Commission items by Project (all commissions for same project are summed)
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

		// Build description: "Bonus for Renaiss" or "Bonus"
		description := "Bonus"
		if project != "" {
			description = fmt.Sprintf("Bonus for %s", project)
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
		})
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

	l.Debug(fmt.Sprintf("found bank account: accountHolder=%s bank=%s accountNumber=%s",
		bankAccount.AccountHolderName, bankAccount.BankName, bankAccount.AccountNumber))

	// 6. Total is already in USD (converted per line item)
	totalUSD := total
	l.Debug(fmt.Sprintf("total USD: %.2f", totalUSD))

	// 7. Generate invoice number
	invoiceNumber := c.generateContractorInvoiceNumber(month)
	l.Debug(fmt.Sprintf("generated invoice number: %s", invoiceNumber))

	// 8. Calculate dates
	now := time.Now()
	dueDate := now.AddDate(0, 0, 15) // Due in 15 days

	// 9. Generate description
	description := fmt.Sprintf("Professional Services for %s", timeutil.FormatMonthYear(month))
	l.Debug(fmt.Sprintf("generated description: %s", description))

	invoiceData := &ContractorInvoiceData{
		InvoiceNumber:     invoiceNumber,
		ContractorName:    rateData.ContractorName,
		Month:             month,
		Date:              now,
		DueDate:           dueDate,
		Description:       description,
		BillingType:       rateData.BillingType,
		Currency:          "USD", // All amounts are converted to USD
		LineItems:         lineItems,
		MonthlyFixed:      0,
		MonthlyFixedUSD:   0,
		Total:             totalUSD,
		TotalUSD:          totalUSD,
		ExchangeRate:      1, // Already in USD
		BankAccountHolder: bankAccount.AccountHolderName,
		BankName:          bankAccount.BankName,
		BankAccountNumber: bankAccount.AccountNumber,
		BankSwiftBIC:      bankAccount.SwiftBIC,
		BankBranch:        bankAccount.BranchAddress,
	}

	l.Debug("contractor invoice data generated successfully")
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
			// Replace bullet points with newlines for better formatting
			formatted := strings.ReplaceAll(text, " • ", "\n• ")
			formatted = strings.ReplaceAll(formatted, " •", "\n•")
			// Replace newlines with <br> for HTML rendering
			formatted = strings.ReplaceAll(formatted, "\n", "<br>")
			return template.HTML(strings.TrimSpace(formatted))
		},
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

	// Prepare template data
	templateData := struct {
		Invoice     *ContractorInvoiceData
		LineItems   []ContractorInvoiceLineItem
		MergedItems []ContractorInvoiceLineItem
		MergedTotal float64
	}{
		Invoice:     data,
		LineItems:   regularItems,
		MergedItems: mergedItems,
		MergedTotal: mergedTotal,
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
