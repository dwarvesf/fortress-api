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
	"strings"
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
}

// GenerateContractorInvoice generates contractor invoice data from Notion
func (c *controller) GenerateContractorInvoice(ctx context.Context, discord, month string) (*ContractorInvoiceData, error) {
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

	l.Debug(fmt.Sprintf("found contractor rate: contractor=%s billingType=%s currency=%s monthlyFixed=%.2f hourlyRate=%.2f",
		rateData.ContractorName, rateData.BillingType, rateData.Currency, rateData.MonthlyFixed, rateData.HourlyRate))

	// 2. Query Task Order Log for the month
	l.Debug("querying task order log from Notion")
	taskOrderService := notion.NewTaskOrderLogService(c.config, c.logger)
	if taskOrderService == nil {
		l.Error(nil, "failed to create task order log service")
		return nil, fmt.Errorf("failed to create task order log service")
	}

	// Check if order exists for this contractor and month
	exists, orderPageID, err := taskOrderService.CheckOrderExistsByContractor(ctx, rateData.ContractorPageID, month)
	if err != nil {
		l.Error(err, "failed to check order existence")
		return nil, fmt.Errorf("failed to check order existence: %w", err)
	}

	if !exists {
		l.Debug("no task order log found for contractor")
		return nil, fmt.Errorf("task order log not found for contractor=%s month=%s", discord, month)
	}

	l.Debug(fmt.Sprintf("found order: pageID=%s", orderPageID))

	// 3. Query Order Subitems
	l.Debug("querying order subitems from Notion")
	subitems, err := taskOrderService.QueryOrderSubitems(ctx, orderPageID)
	if err != nil {
		l.Error(err, "failed to query order subitems")
		return nil, fmt.Errorf("failed to query order subitems: %w", err)
	}

	if len(subitems) == 0 {
		l.Debug("no subitems found for order")
		return nil, fmt.Errorf("no line items found in task order log for order=%s", orderPageID)
	}

	l.Debug(fmt.Sprintf("found %d subitems", len(subitems)))

	// 4. Build line items based on billing type
	var lineItems []ContractorInvoiceLineItem
	var total float64

	if rateData.BillingType == "Monthly Fixed" {
		l.Debug("building line items for Monthly Fixed billing")
		for _, subitem := range subitems {
			lineItems = append(lineItems, ContractorInvoiceLineItem{
				Title:       subitem.ProjectName,
				Description: subitem.ProofOfWork,
			})
		}
		total = rateData.MonthlyFixed
	} else if rateData.BillingType == "Hourly Rate" {
		l.Debug("building line items for Hourly Rate billing")
		for _, subitem := range subitems {
			amount := subitem.Hours * rateData.HourlyRate
			// Convert line item amount to USD
			amountUSD, _, err := c.service.Wise.Convert(amount, rateData.Currency, "USD")
			if err != nil {
				l.Error(err, "failed to convert line item amount to USD")
				amountUSD = 0 // Fallback to 0 if conversion fails
			}
			lineItems = append(lineItems, ContractorInvoiceLineItem{
				Title:       subitem.ProjectName,
				Description: subitem.ProofOfWork,
				Hours:       subitem.Hours,
				Rate:        rateData.HourlyRate,
				Amount:      amount,
				AmountUSD:   amountUSD,
			})
			total += amount
		}
	} else {
		l.Error(nil, fmt.Sprintf("unsupported billing type: %s", rateData.BillingType))
		return nil, fmt.Errorf("unsupported billing type: %s", rateData.BillingType)
	}

	l.Debug(fmt.Sprintf("built %d line items with total=%.2f", len(lineItems), total))

	// 5. Query Bank Account
	l.Debug("querying bank account from Notion")
	bankAccountService := notion.NewBankAccountService(c.config, c.logger)
	if bankAccountService == nil {
		l.Error(nil, "failed to create bank account service")
		return nil, fmt.Errorf("failed to create bank account service")
	}

	bankAccount, err := bankAccountService.QueryBankAccountByDiscord(ctx, discord)
	if err != nil {
		l.Error(err, "failed to query bank account")
		return nil, fmt.Errorf("bank account not found: %w", err)
	}

	l.Debug(fmt.Sprintf("found bank account: accountHolder=%s bank=%s accountNumber=%s",
		bankAccount.AccountHolderName, bankAccount.BankName, bankAccount.AccountNumber))

	// 6. Convert to USD using Wise
	monthlyFixedUSD, _, err := c.service.Wise.Convert(rateData.MonthlyFixed, rateData.Currency, "USD")
	if err != nil {
		l.Error(err, "failed to convert monthly fixed amount to USD")
		monthlyFixedUSD = 0 // Fallback to 0 if conversion fails
	}

	// 6.1 Convert to USD using Wise
	l.Debug("converting total to USD using Wise")
	totalUSD, exchangeRate, err := c.service.Wise.Convert(total, rateData.Currency, "USD")
	if err != nil {
		l.Error(err, "failed to convert to USD")
		return nil, fmt.Errorf("failed to convert to USD: %w", err)
	}
	l.Debug(fmt.Sprintf("converted total: %.2f %s = %.2f USD (rate: %.6f)", total, rateData.Currency, totalUSD, exchangeRate))

	// 7. Generate invoice number
	invoiceNumber := c.generateContractorInvoiceNumber(month)
	l.Debug(fmt.Sprintf("generated invoice number: %s", invoiceNumber))

	// 8. Calculate dates
	now := time.Now()
	dueDate := now.AddDate(0, 0, 15) // Due in 15 days

	// 9. Generate description
	description := fmt.Sprintf("Software Development Services for %s", timeutil.FormatMonthYear(month))
	l.Debug(fmt.Sprintf("generated description: %s", description))

	invoiceData := &ContractorInvoiceData{
		InvoiceNumber:     invoiceNumber,
		ContractorName:    rateData.ContractorName,
		Month:             month,
		Date:              now,
		DueDate:           dueDate,
		Description:       description,
		BillingType:       rateData.BillingType,
		Currency:          rateData.Currency,
		LineItems:         lineItems,
		MonthlyFixed:      rateData.MonthlyFixed,
		MonthlyFixedUSD:   monthlyFixedUSD,
		Total:             total,
		TotalUSD:          totalUSD,
		ExchangeRate:      exchangeRate,
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
			return pound.Multiply(int64(tmpValue)).Display()
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
		"formatProofOfWork": func(text string) string {
			// Replace bullet points with newlines for better formatting
			formatted := strings.ReplaceAll(text, " • ", "\n• ")
			formatted = strings.ReplaceAll(formatted, " •", "\n•")
			return strings.TrimSpace(formatted)
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

	// Separate line items into regular and merged (no hours/rate)
	var regularItems []ContractorInvoiceLineItem
	var mergedItems []ContractorInvoiceLineItem

	for _, item := range data.LineItems {
		// Only merge items with Hours, Rate, and Amount all equal to 0
		// Items with MonthlyFixed (Amount > 0) should remain as regular items
		if item.Hours == 0 && item.Rate == 0 && item.Amount == 0 {
			mergedItems = append(mergedItems, item)
		} else {
			regularItems = append(regularItems, item)
		}
	}

	// Unit cost and total for merged items = MonthlyFixedUSD
	mergedTotal := data.MonthlyFixedUSD

	l.Debug(fmt.Sprintf("separated line items: %d regular, %d merged (total=%.2f)",
		len(regularItems), len(mergedItems), mergedTotal))

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
