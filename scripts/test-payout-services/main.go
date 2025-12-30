package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/controller/invoice"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

func main() {
	cfg := config.LoadConfig(config.DefaultConfigLoaders())
	l := logger.NewLogrusLogger("debug")

	// Support both CONTRACTOR_PAGE_ID and CONTRACTOR_DISCORD
	contractorPageID := os.Getenv("CONTRACTOR_PAGE_ID")
	contractorDiscord := os.Getenv("CONTRACTOR_DISCORD")
	month := os.Getenv("MONTH")
	generatePDF := os.Getenv("GENERATE_PDF") == "true"

	if contractorPageID == "" && contractorDiscord == "" {
		l.Error(nil, "Either CONTRACTOR_PAGE_ID or CONTRACTOR_DISCORD env var is required")
		os.Exit(1)
	}

	if month == "" {
		month = time.Now().Format("2006-01")
		l.Debugf("Using current month: %s", month)
	}

	ctx := context.Background()

	// If Discord provided, look up contractor page ID
	contractorName := contractorDiscord
	if contractorPageID == "" && contractorDiscord != "" {
		l.Debugf("Looking up contractor by Discord: %s", contractorDiscord)
		pageID, name, err := lookupContractorByDiscord(ctx, cfg, l, contractorDiscord)
		if err != nil {
			l.Error(err, "Failed to lookup contractor by Discord")
			os.Exit(1)
		}
		contractorPageID = pageID
		if name != "" {
			contractorName = name
		}
	}

	l.Debugf("Testing with contractor page ID: %s", contractorPageID)

	payoutsSvc := notion.NewContractorPayoutsService(cfg, l)
	feesSvc := notion.NewContractorFeesService(cfg, l)
	splitSvc := notion.NewInvoiceSplitService(cfg, l)
	taskOrderLogSvc := notion.NewTaskOrderLogService(cfg, l)
	wiseSvc := wise.New(cfg, l)

	// Test QueryPendingPayoutsByContractor
	l.Debug("Querying pending payouts...")
	entries, err := payoutsSvc.QueryPendingPayoutsByContractor(ctx, contractorPageID)
	if err != nil {
		l.Error(err, "QueryPendingPayoutsByContractor failed")
		os.Exit(1)
	}

	fmt.Printf("Found %d pending payouts\n", len(entries))

	// Build line items for PDF
	var lineItems []invoice.ContractorInvoiceLineItem
	var totalUSD float64

	for _, e := range entries {
		fmt.Printf("- %s: %s %.2f %s (Name: %s)\n", e.SourceType, e.Direction, e.Amount, e.Currency, e.Name)

		// Convert amount to USD and round to 2 decimal places
		amountUSD := e.Amount
		if e.Currency != "" && e.Currency != "USD" {
			l.Debugf("Converting %.2f %s to USD", e.Amount, e.Currency)
			converted, rate, err := wiseSvc.Convert(e.Amount, e.Currency, "USD")
			if err != nil {
				l.Debugf("Failed to convert currency: %v (using original amount)", err)
			} else {
				amountUSD = converted
				l.Debugf("Converted: %.2f %s = %.2f USD (rate: %.4f)", e.Amount, e.Currency, amountUSD, rate)
			}
		}
		// Round to 2 decimal places to avoid $0.01 differences
		amountUSD = roundTo2Decimals(amountUSD)

		// Initialize line item with defaults (Qty=1, Rate=AmountUSD, Total=AmountUSD)
		lineItem := invoice.ContractorInvoiceLineItem{
			Title:     e.Name,
			Hours:     1,
			Rate:      amountUSD,
			Amount:    amountUSD,
			AmountUSD: amountUSD,
		}

		// For Contractor Payroll, fetch ProofOfWorks from Task Order Log subitems (grouped by project)
		if e.ContractorFeesID != "" && taskOrderLogSvc != nil {
			l.Debugf("Fetching Task Order Log IDs from contractor fees: %s", e.ContractorFeesID)

			// Get Task Order Log IDs from Contractor Fees relation
			orderIDs, err := feesSvc.GetTaskOrderLogIDs(ctx, e.ContractorFeesID)
			if err != nil {
				l.Error(err, "GetTaskOrderLogIDs failed")
			} else if len(orderIDs) > 0 {
				l.Debugf("Found %d Task Order Log IDs: %v", len(orderIDs), orderIDs)

				// Format ProofOfWorks grouped by project with bold headers
				formattedDescription, err := taskOrderLogSvc.FormatProofOfWorksByProject(ctx, orderIDs)
				if err != nil {
					l.Error(err, "FormatProofOfWorksByProject failed")
				} else if formattedDescription != "" {
					lineItem.Description = formattedDescription
					fmt.Printf("  Formatted ProofOfWorks:\n%s\n", formattedDescription)
					l.Debugf("Set formatted ProofOfWorks description (length=%d)", len(formattedDescription))
				} else {
					l.Debug("No ProofOfWorks found in Task Order Log subitems")
				}
			} else {
				l.Debug("No Task Order Log IDs found in contractor fees")
			}
		}

		// Test GetInvoiceSplitByID if split relation exists
		if e.InvoiceSplitID != "" {
			l.Debugf("Fetching invoice split: %s", e.InvoiceSplitID)
			split, err := splitSvc.GetInvoiceSplitByID(ctx, e.InvoiceSplitID)
			if err != nil {
				l.Error(err, "GetInvoiceSplitByID failed")
			} else {
				fmt.Printf("  Split: %s - %.2f %s\n", split.Role, split.Amount, split.Currency)
			}
		}

		lineItems = append(lineItems, lineItem)
		totalUSD += lineItem.AmountUSD
	}

	l.Info("Test completed successfully")

	// Generate PDF if requested
	if generatePDF {
		l.Debug("Generating invoice PDF...")

		// Create controller for PDF generation
		ctrl := controller.New(nil, nil, nil, nil, l, cfg)

		// Build invoice data
		invoiceData := &invoice.ContractorInvoiceData{
			InvoiceNumber:     generateInvoiceNumber(month),
			ContractorName:    contractorName,
			Month:             month,
			Date:              time.Now(),
			DueDate:           time.Now().AddDate(0, 0, 15),
			Description:       fmt.Sprintf("Software Development Services for %s", timeutil.FormatMonthYear(month)),
			BillingType:       "Hourly Rate",
			Currency:          "USD",
			LineItems:         lineItems,
			Total:             totalUSD,
			TotalUSD:          totalUSD,
			ExchangeRate:      1,
			BankAccountHolder: "Test Account Holder",
			BankName:          "Test Bank",
			BankAccountNumber: "123456789",
			BankSwiftBIC:      "TESTSWIFT",
			BankBranch:        "Test Branch",
		}

		// Try to get bank account from Notion
		bankSvc := notion.NewBankAccountService(cfg, l)
		if bankSvc != nil && contractorDiscord != "" {
			l.Debugf("Fetching bank account for Discord: %s", contractorDiscord)
			bankAccount, err := bankSvc.QueryBankAccountByDiscord(ctx, contractorDiscord)
			if err != nil {
				l.Debugf("Failed to fetch bank account: %v (using defaults)", err)
			} else {
				invoiceData.BankAccountHolder = bankAccount.AccountHolderName
				invoiceData.BankName = bankAccount.BankName
				invoiceData.BankAccountNumber = bankAccount.AccountNumber
				invoiceData.BankSwiftBIC = bankAccount.SwiftBIC
				invoiceData.BankBranch = bankAccount.BranchAddress
				l.Debugf("Using bank account: %s at %s", bankAccount.AccountHolderName, bankAccount.BankName)
			}
		}

		pdfBytes, err := ctrl.Invoice.GenerateContractorInvoicePDF(l, invoiceData)
		if err != nil {
			l.Error(err, "Failed to generate PDF")
			os.Exit(1)
		}

		outputFile := fmt.Sprintf("contractor-invoice-%s-%s.pdf", sanitizeFilename(contractorName), month)
		if err := os.WriteFile(outputFile, pdfBytes, 0644); err != nil {
			l.Error(err, "Failed to write PDF file")
			os.Exit(1)
		}
		l.Info(fmt.Sprintf("PDF generated: %s (%d bytes)", outputFile, len(pdfBytes)))
	}
}

func lookupContractorByDiscord(ctx context.Context, cfg *config.Config, l logger.Logger, discord string) (string, string, error) {
	client := nt.NewClient(cfg.Notion.Secret)
	contractorDBID := cfg.Notion.Databases.Contractor

	l.Debugf("Querying contractor database: %s for Discord: %s", contractorDBID, discord)

	filter := &nt.DatabaseQueryFilter{
		Property: "Discord",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			RichText: &nt.TextPropertyFilter{
				Equals: discord,
			},
		},
	}

	query := &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 1,
	}

	resp, err := client.QueryDatabase(ctx, contractorDBID, query)
	if err != nil {
		return "", "", fmt.Errorf("query contractor database failed: %w", err)
	}

	if len(resp.Results) == 0 {
		return "", "", fmt.Errorf("no contractor found with Discord: %s", discord)
	}

	page := resp.Results[0]
	pageID := page.ID

	// Extract contractor name
	var name string
	if props, ok := page.Properties.(nt.DatabasePageProperties); ok {
		if nameProp, ok := props["Name"]; ok && len(nameProp.Title) > 0 {
			for _, rt := range nameProp.Title {
				name += rt.PlainText
			}
		}
	}

	l.Debugf("Found contractor page ID: %s, Name: %s", pageID, name)

	return pageID, name, nil
}

func generateInvoiceNumber(month string) string {
	monthPart := strings.ReplaceAll(month, "-", "")
	return fmt.Sprintf("CONTR-%s-TEST", monthPart)
}

func sanitizeFilename(name string) string {
	// Replace spaces and special characters with underscores
	result := strings.ReplaceAll(name, " ", "_")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.ReplaceAll(result, "\\", "_")
	return result
}

func roundTo2Decimals(val float64) float64 {
	return math.Round(val*100) / 100
}
